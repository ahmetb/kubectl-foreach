// Copyright 2022 Twitter, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/jwalton/gchalk"
	"golang.org/x/sync/errgroup"
)

const (
	envDisablePrompts = `ALLCTX_DISABLE_PROMPTS`
)

var (
	chalk = gchalk.Stderr
	gray  = chalk.Gray
	red   = chalk.Red

	repl    = flag.String("I", "", "string to replace in cmd args with context name (like xargs -I)")
	workers = flag.Int("c", 0, "parallel runs (default: as many as matched contexts)")
	quiet   = flag.Bool("q", false, "accept confirmation prompts")
)

func printErrAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "%s", red("error: "))
	os.Exit(1)
}

func main() {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	// initialize signal handler after
	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, gray("received exit signal"))
	}()

	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	flag.Parse()
	if *workers < 0 {
		printErrAndExit("-c < 0")
	}
	filters, args, err := parseArgs(os.Args[1:])
	if err != nil {
		printErrAndExit(fmt.Errorf("failed to parse command-line arguments: %w", err).Error())
	}

	ctxs, err := kubeContexts(ctx)
	if err != nil {
		printErrAndExit(err.Error())
	}
	ctxMatches := matchContexts(ctxs, filters)

	if len(ctxMatches) == 0 {
		printErrAndExit("query matched no contexts from kubeconfig")
	}

	if os.Getenv(envDisablePrompts) == "" {
		if *quiet {
			for _, c := range ctxMatches {
				fmt.Fprintf(os.Stderr, "%s", gray(fmt.Sprintf("  - %s\n", c)))
			}
		} else {
			fmt.Fprintln(os.Stderr, "Will run command in context(s):")
			for _, c := range ctxMatches {
				fmt.Fprintf(os.Stderr, "%s", gray(fmt.Sprintf("  - %s\n", c)))
			}
			fmt.Fprintf(os.Stderr, "Continue? [Y/n]: ")
			if err := prompt(ctx, os.Stdin); err != nil {
				printErrAndExit(err.Error())
			}
		}
	}

	syncOut := &synchronizedWriter{Writer: os.Stdout}
	syncErr := &synchronizedWriter{Writer: os.Stderr}

	err = runAll(ctx, ctxMatches, replaceArgs(args, *repl), syncOut, syncErr)
	if err != nil {
		printErrAndExit(err.Error())
	}
}

func replaceArgs(args []string, repl string) func(ctx string) []string {
	return func(ctx string) []string {
		if repl == "" {
			return append([]string{"--context=" + ctx}, args...)
		}
		out := make([]string, len(args))
		for i := range args {
			out[i] = strings.Replace(args[i], repl, ctx, -1)
		}
		return out
	}
}

func kubeContexts(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "get-contexts", "-o=name")
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = os.Stderr // TODO might be redundant
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get contexts: %w", err)
	}
	return strings.Split(strings.TrimSpace(b.String()), "\n"), nil
}

func runAll(ctx context.Context, kubeCtxs []string, argMaker func(string) []string, stdout, stderr io.Writer) error {
	n := len(kubeCtxs)
	if *workers > 0 {
		n = *workers
	}

	wg, _ := errgroup.WithContext(ctx)
	wg.SetLimit(n)

	maxLen := maxLen(kubeCtxs)
	leftPad := func(s string, origLen int) string {
		return strings.Repeat(" ", maxLen-origLen) + s
	}

	colors := []func(string, ...interface{}) string{
		// foreground only
		chalk.WithRed().Sprintf,
		chalk.WithBlue().Sprintf,
		chalk.WithGreen().Sprintf,
		chalk.WithYellow().WithBgBlack().Sprintf,
		chalk.WithGray().Sprintf,
		chalk.WithMagenta().Sprintf,
		chalk.WithCyan().Sprintf,
		chalk.WithBrightRed().Sprintf,

		chalk.WithBrightBlue().Sprintf,
		chalk.WithBrightGreen().Sprintf,
		chalk.WithBrightMagenta().Sprintf,
		chalk.WithBrightYellow().WithBgBlack().Sprintf,
		chalk.WithBrightCyan().Sprintf,

		// inverse
		chalk.WithBgRed().WithWhite().Sprintf,
		chalk.WithBgBlue().WithWhite().Sprintf,
		chalk.WithBgCyan().WithBlack().Sprintf,
		chalk.WithBgGreen().WithBlack().Sprintf,
		chalk.WithBgMagenta().WithBrightWhite().Sprintf,
		chalk.WithBgYellow().WithBlack().Sprintf,
		chalk.WithBgGray().WithWhite().Sprintf,
		chalk.WithBgBrightRed().WithWhite().Sprintf,
		chalk.WithBgBrightBlue().WithWhite().Sprintf,
		chalk.WithBgBrightCyan().WithBlack().Sprintf,
		chalk.WithBgBrightGreen().WithBlack().Sprintf,
		chalk.WithBgBrightMagenta().WithBlack().Sprintf,
		chalk.WithBgBrightYellow().WithBlack().Sprintf,

		// mixes+inverses
		chalk.WithBgRed().WithYellow().Sprintf,
		chalk.WithBgYellow().WithRed().Sprintf,
		chalk.WithBgBlue().WithYellow().Sprintf,
		chalk.WithBgYellow().WithBlue().Sprintf,
		chalk.WithBgBlack().WithBrightWhite().Sprintf,
		chalk.WithBgBrightWhite().WithBlack().Sprintf,
	}

	for i, kctx := range kubeCtxs {
		kctx := kctx
		ctx := ctx
		colFn := colors[i%len(colors)]
		wg.Go(func() error {
			prefix := []byte(leftPad(colFn(kctx), len(kctx)) + " | ")
			wo := &prefixingWriter{prefix: prefix, w: stdout}
			we := &prefixingWriter{prefix: prefix, w: stderr}
			return run(ctx, argMaker(kctx), wo, we)
		})
	}
	return wg.Wait()
}

func maxLen(s []string) int {
	max := 0
	for _, v := range s {
		if len(v) > max {
			max = len(v)
		}
	}
	return max
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) (err error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func prompt(ctx context.Context, r io.Reader) error {
	pr, pw := io.Pipe()
	go io.Copy(pw, r)
	defer pw.Close()

	scanDone := make(chan error)

	go func() {
		s := bufio.NewScanner(pr)
		for s.Scan() {
			v := s.Text()
			if v == "y" || v == "Y" || v == "" {
				scanDone <- nil
			}
			break
		}
		scanDone <- errors.New("user refused execution")
	}()

	select {
	case res := <-scanDone:
		return res
	case <-ctx.Done():
		pr.Close()
		return fmt.Errorf("prompt canceled")
	}
}

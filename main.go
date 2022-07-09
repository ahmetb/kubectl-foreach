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
	"regexp"
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
	auto    = false
)

func logErr(msg string) {
	fmt.Fprintf(os.Stderr, "%s", red("error: "))
	fmt.Fprintf(os.Stderr, "%v\n", msg)
	os.Exit(1)
}

func main() {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, gray("received exit signal"))
	}()

	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	flag.Parse()
	if *workers < 0 {
		logErr("-c < 0")
	}

	var filters []filter
	var args []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "~") {
			re, err := regexp.Compile(arg[1:])
			if err != nil {
				logErr(fmt.Sprintf("bad pattern: %v", err))
			}
			filters = append(filters, pattern{re})
		} else {
			filters = append(filters, exact(arg))
		}

		if i+1 == len(os.Args) {
			// end
			if arg == "--" {
				logErr("need more args after '--'")
			} else if arg == "--auto" {
				logErr("need more args after '--auto'")
			} else {
				logErr("did not find '--' as an argument, see -h")
			}
		}
		if arg == "--" {
			args = os.Args[i:]
			break
		} else if arg == "--auto" {
			args = os.Args[i:]
			auto = true
			break
		}
	}
	if len(args) == 0 {
		logErr("must supply arguments/options to kubectl after '--' or '--auto'")
	}
	args = args[1:]

	ctxs, err := kubeContexts(ctx)
	if err != nil {
		logErr(err.Error())
	}
	var outCtx []string
	for _, c := range ctxs {
		for _, f := range filters {
			if f.match(c) {
				outCtx = append(outCtx, c)
				break
			}
		}
	}

	if len(outCtx) == 0 {
		logErr("query matched no contexts from kubeconfig")
	}

	if os.Getenv(envDisablePrompts) == "" {
		if auto {
			fmt.Fprintln(os.Stderr, "Running command in context(s) automatically:")
			for _, c := range outCtx {
				fmt.Fprintf(os.Stderr, "%s", gray(fmt.Sprintf("  - %s\n", c)))
			}
		} else {
			fmt.Fprintln(os.Stderr, "Will run command in context(s):")
			for _, c := range outCtx {
				fmt.Fprintf(os.Stderr, "%s", gray(fmt.Sprintf("  - %s\n", c)))
			}
			fmt.Fprintf(os.Stderr, "Continue? [Y/n]: ")
			if err := prompt(os.Stdin); err != nil {
				logErr(err.Error())
			}
		}
	}

	syncOut := &synchronizedWriter{Writer: os.Stdout}
	syncErr := &synchronizedWriter{Writer: os.Stderr}

	err = runAll(ctx, outCtx, replaceArgs(args, *repl), syncOut, syncErr)
	if err != nil {
		logErr(err.Error())
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

type filter interface {
	match(string) bool
}

type exact string

func (e exact) match(in string) bool {
	return in == string(e)
}

type pattern struct{ *regexp.Regexp }

func (p pattern) match(in string) bool {
	return p.MatchString(in)
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

func prompt(r io.Reader) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		v := s.Text()
		if v == "y" || v == "Y" || v == "" {
			return nil
		}
		break
	}
	return errors.New("cancelled")
}

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
	envDisablePrompts = `KUBECTL_FOREACH_DISABLE_PROMPTS`
)

var (
	gray = gchalk.Stderr.Gray
	red  = gchalk.Stderr.Red

	fl      = flag.NewFlagSet("kubectl foreach", flag.ContinueOnError)
	repl    = fl.String("I", "", "string to replace in cmd args with context name (like xargs -I)")
	workers = fl.Int("c", 0, "parallel runs (default: as many as matched contexts)")
	quiet   = fl.Bool("q", false, "accept confirmation prompts")
)

func printErrAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "%s%s\n", red("error: "), msg)
	os.Exit(1)
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `Usage:
    kubectl foreach [OPTIONS] [PATTERN]... -- [KUBECTL_ARGS...]

Patterns can be used to match context names from kubeconfig:
      (empty): matches all contexts
         NAME: matches context with exact name
    /PATTERN/: matches context with regular expression
        ^NAME: remove context with exact name from the matched results
   ^/PATTERN/: remove contexts matching the regular expression from the results
    
Options:
    -c=NUM     Limit parallel executions (default: 0, unlimited)
    -I=VAL     Replace VAL occurring in KUBECTL_ARGS with context name
    -q         Disable and accept confirmation prompts ($KUBECTL_FOREACH_DISABLE_PROMPTS) 
    -h/--help  Print help

Examples:
    # get nodes on contexts named a b c
    kubectl foreach a b c -- get nodes 

    # get nodes on all contexts named c0..9 except c1 (note the escaping)
    kubectl foreach '/^c[0-9]/' ^c1	 -- get nodes

    # get nodes on all contexts that has "prod" but not "foo"
    kubectl foreach /prod/ ^/foo/ -- get nodes

    # use 'kubectl tail' plugin to follow logs of pods in contexts named *test*
    kubectl foreach -I _ /test/ -- tail --context=_ -l app=foo`+"\n")
	os.Exit(0)
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	fl.Usage = func() { printUsage(os.Stderr) }

	if err := fl.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(os.Stderr)
		}
		printErrAndExit(err.Error())
	}
	_, kubectlArgs, err := separateArgs(os.Args[1:])
	if err != nil {
		printErrAndExit(fmt.Errorf("failed to parse command-line arguments: %w. see -h/--help", err).Error())
	}

	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)
	// initialize signal handler after
	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, gray("received exit signal"))
	}()

	if *workers < 0 {
		printErrAndExit("-c < 0")
	}

	ctxs, err := kubeContexts(ctx)
	if err != nil {
		printErrAndExit(err.Error())
	}
	var filters []filter

	// re-parse flags to extract positional arguments of the tool, minus '--' + kubectl args
	if err := fl.Parse(trimSuffix(os.Args[1:], append([]string{"--"}, kubectlArgs...))); err != nil {
		printErrAndExit(err.Error())
	}
	for _, arg := range fl.Args() {
		f, err := parseFilter(arg)
		if err != nil {
			printErrAndExit(err.Error())
		}
		filters = append(filters, f)
	}

	ctxMatches := matchContexts(ctxs, filters)

	if len(ctxMatches) == 0 {
		printErrAndExit("query matched no contexts from kubeconfig")
	}

	fmt.Fprintln(os.Stderr, "Will run command in context(s):")
	for _, c := range ctxMatches {
		fmt.Fprintf(os.Stderr, "%s", gray(fmt.Sprintf("  - %s\n", c)))
	}
	if !*quiet && os.Getenv(envDisablePrompts) == "" {
		fmt.Fprintf(os.Stderr, "Continue? [Y/n]: ")
		if err := prompt(ctx, os.Stdin); err != nil {
			printErrAndExit(err.Error())
		}
	}

	syncOut := &synchronizedWriter{Writer: os.Stdout}
	syncErr := &synchronizedWriter{Writer: os.Stderr}

	err = runAll(ctx, ctxMatches, replaceArgs(kubectlArgs, *repl), syncOut, syncErr)
	if err != nil {
		printErrAndExit(err.Error())
	}
}

func replaceArgs(args []string, repl string) func(ctx string) ([]string, error) {
	return func(ctx string) ([]string, error) {
		if repl == "" {
			return append([]string{"--context=" + ctx}, args...), nil
		}
		out := make([]string, len(args))
		modified := false
		for i := range args {
			out[i] = strings.Replace(args[i], repl, ctx, -1)
			modified = modified || (out[i] != args[i])
		}
		if !modified {
			return nil, fmt.Errorf("args did not use context name replacement string %q", repl)
		}
		return out, nil
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

func runAll(ctx context.Context, kubeCtxs []string, argMaker func(string) ([]string, error), stdout, stderr io.Writer) error {
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

	for i, kctx := range kubeCtxs {
		kctx := kctx
		ctx := ctx
		i := i
		colFn := colors[i%len(colors)]
		wg.Go(func() error {
			prefix := []byte(leftPad(colFn(kctx), len(kctx)) + " | ")
			wo := &prefixingWriter{prefix: prefix, w: stdout}
			we := &prefixingWriter{prefix: prefix, w: stderr}
			args, err := argMaker(kctx)
			if err != nil {
				return err
			}
			return run(ctx, args, wo, we)
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

func run(ctx context.Context, args []string, stdout, stderr io.WriteCloser) (err error) {
	defer func() {
		// flush underlying writer (prefixWriter) by closing in case last output does not terminate with newline
		if err := stdout.Close(); err != nil {
			log.Printf("WARN: failed to close stdout: %v", err)
		}
		if err := stderr.Close(); err != nil {
			log.Printf("WARN: failed to close stdout: %v", err)
		}
	}()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// prompt returns an error if user rejects or if ctx cancels.
func prompt(ctx context.Context, r io.Reader) error {
	pr, pw := io.Pipe()
	go func() {
		if _, err := io.Copy(pw, r); err != nil {
			pw.Close()
		}
	}()
	defer pw.Close()

	scanDone := make(chan error, 1)

	go func() {
		s := bufio.NewScanner(pr)
		for s.Scan() {
			if err := s.Err(); err != nil {
				scanDone <- err
				return
			}
			v := s.Text()
			if v == "y" || v == "Y" || v == "" {
				scanDone <- nil
				return
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

func trimSuffix(a []string, suffix []string) []string {
	if len(suffix) > len(a) {
		return a
	}
	for i, j := len(a)-1, len(suffix)-1; j >= 0; i, j = i-1, j-1 {
		if a[i] != suffix[j] {
			return a
		}
	}
	return a[:len(a)-len(suffix)]
}

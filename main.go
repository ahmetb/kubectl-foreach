package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

const (
	envDisablePrompts = `ALLCTX_DISABLE_PROMPTS`
)

var (
	gray = color.New(color.FgHiBlack)
	red  = color.New(color.FgRed)
)

func logErr(msg string) {
	red.Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, "%v\n", msg)
	os.Exit(1)
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	flag.Parse()

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
			} else {
				logErr("did not find '--' as an argument, see -h")
			}
		}
		if arg == "--" {
			args = os.Args[i:]
			break
		}
	}
	if len(args) == 0 {
		logErr("must supply arguments/options to kubectl after '--'")
	}
	args = args[1:]

	ctxs, err := contexts()
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
		red.Fprintf(os.Stderr, "query matched no contexts from kubeconfig")
	}

	if os.Getenv(envDisablePrompts) == "" {
		fmt.Fprintln(os.Stderr, "Will run command in contexts:")
		for _, c := range outCtx {
			gray.Fprintf(os.Stderr, "  - %s\n", c)
		}
		fmt.Fprintf(os.Stderr, "Continue? [Y/n]: ")
		if err := prompt(os.Stdin); err != nil {
			red.Fprintf(os.Stderr, "error: ")
			log.Fatal(err)
		}
	}

	syncOut := &synchronizedWriter{Writer: os.Stdout}
	syncErr := &synchronizedWriter{Writer: os.Stderr}

	err = runAll(outCtx, args, syncOut, syncErr)
	if err != nil {
		logErr(err.Error())
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

func contexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o=name")
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = os.Stderr // TODO might be redundant
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get contexts: %w", err)
	}
	return strings.Split(b.String(), "\n"), nil
}

func runAll(ctxs []string, args []string, stdout, stderr io.Writer) error {

	var wg errgroup.Group

	maxLen := maxLen(ctxs)
	leftPad := func(s string) string {
		return strings.Repeat(" ", maxLen-len(s)) + s
	}

	colors := []func(string, ...interface{}) string{
		color.RedString,
		color.CyanString,
		color.GreenString,
		color.MagentaString,
		color.YellowString,
		color.HiWhiteString,
		color.HiBlackString,
		color.HiRedString,
		color.HiCyanString,
		color.HiGreenString,
		color.HiMagentaString,
		color.HiYellowString,
	}

	for i, ctx := range ctxs {
		ctx := ctx
		colFn := colors[i%len(colors)]
		wg.Go(func() error {
			prefix := []byte(colFn(leftPad(ctx)) + " | ")
			wo := &prefixingWriter{prefix: prefix, w: stdout}
			we := &prefixingWriter{prefix: prefix, w: stderr}
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

func run(ctx string, args []string, stdout, stderr io.Writer) (err error) {
	cmd := exec.Command("kubectl", append([]string{"--context=" + ctx}, args...)...)
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

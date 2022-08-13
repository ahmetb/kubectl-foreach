package main

import (
	"errors"
	"fmt"
	"regexp"
)

type filter interface {
	match(string) bool
	additive() bool
}

type exact string

func (e exact) match(in string) bool { return in == string(e) }
func (exact) additive() bool         { return true }

type pattern struct{ *regexp.Regexp }

func (p pattern) match(in string) bool { return p.MatchString(in) }
func (pattern) additive() bool         { return true }

type exclude struct{ filter }

func (e exclude) match(s string) bool { return e.filter.match(s) }
func (exclude) additive() bool        { return false }

// parseFilter parses a command-line syntax of a matcher.
func parseFilter(in string) (filter, error) {
	if in == "" {
		return nil, errors.New("empty string cannot be used as a filter")
	}

	var exclusion bool
	if in[0] == '^' {
		in = in[1:]
		exclusion = true
	}
	var f filter
	// pattern /re/
	if len(in) > 1 && in[0] == '/' && in[len(in)-1] == '/' {
		r, err := regexp.Compile(in[1 : len(in)-1])
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", in, err)
		}
		f = pattern{r}
	} else {
		// exact match
		f = exact(in)
	}

	if exclusion {
		return exclude{f}, nil
	}
	return f, nil
}

package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		_, _, err := parseArgs(nil)
		assert.NotNil(t, err)
	})
	t.Run("only argv[1]", func(t *testing.T) {
		_, _, err := parseArgs([]string{})
		assert.NotNil(t, err)
	})
	t.Run("no separator", func(t *testing.T) {
		_, _, err := parseArgs([]string{"a", "b"})
		assert.NotNil(t, err)
	})
	t.Run("only separator", func(t *testing.T) {
		_, _, err := parseArgs([]string{"--"})
		assert.NotNil(t, err)
	})
	t.Run("no command", func(t *testing.T) {
		_, _, err := parseArgs([]string{"a", "b", "--"})
		assert.NotNil(t, err)
	})
	t.Run("filter parse err", func(t *testing.T) {
		_, _, err := parseArgs([]string{"/re[A-Z/", "--", "foo"})
		assert.NotNil(t, err)
	})
	t.Run("parse ok", func(t *testing.T) {
		f, a, err := parseArgs([]string{"/re/", "a", "--", "foo"})
		assert.Nil(t, err)
		assert.Equal(t, []filter{
			pattern{regexp.MustCompile("re")},
			exact("a")}, f)
		assert.Equal(t, []string{"foo"}, a)
	})
}

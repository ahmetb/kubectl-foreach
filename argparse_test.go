package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeparateArgs(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		_, _, err := separateArgs(nil)
		assert.NotNil(t, err)
	})
	t.Run("no separator", func(t *testing.T) {
		_, _, err := separateArgs([]string{"a", "b"})
		assert.NotNil(t, err)
	})
	t.Run("only separator", func(t *testing.T) {
		_, _, err := separateArgs([]string{"--"})
		assert.NotNil(t, err)
	})
	t.Run("no right", func(t *testing.T) {
		_, _, err := separateArgs([]string{"a", "b", "--"})
		assert.NotNil(t, err)
	})
	t.Run("no left", func(t *testing.T) {
		l, r, err := separateArgs([]string{"--", "a", "b"})
		assert.Nil(t, err)
		assert.Empty(t, l)
		assert.Equal(t, []string{"a", "b"}, r)
	})
	t.Run("parses left and right", func(t *testing.T) {
		l, r, err := separateArgs([]string{"a", "b", "--", "foo"})
		assert.Nil(t, err)
		assert.Equal(t, []string{"a", "b"}, l)
		assert.Equal(t, []string{"foo"}, r)
	})
}

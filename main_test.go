package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimSuffix(t *testing.T) {
	assert.Equal(t, []string(nil), trimSuffix(nil, []string{"b"}))
	assert.Equal(t, []string{"a"}, trimSuffix([]string{"a"}, nil))
	assert.Equal(t, []string{"a"}, trimSuffix([]string{"a"}, []string{}))
	assert.Equal(t, []string{}, trimSuffix([]string{"a"}, []string{"a"}))
	assert.Equal(t, []string{}, trimSuffix([]string{"a", "b"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a"}, trimSuffix([]string{"a"}, []string{"b"}))
	assert.Equal(t, []string{"a"}, trimSuffix([]string{"a"}, []string{"a", "b"}))
	assert.Equal(t, []string{"a", "b"}, trimSuffix([]string{"a", "b", "c"}, []string{"c"}))
	assert.Equal(t, []string{"a"}, trimSuffix([]string{"a", "b", "c"}, []string{"b", "c"}))

}

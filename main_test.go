// Copyright 2022 Twitter, Inc.
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
	"context"
	"errors"
	"strings"
	"testing"
	"testing/iotest"
	"time"

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

func TestPrompt(t *testing.T) {
	t.Run("ctx cancel", func(t *testing.T) {
		ch := make(chan struct{})
		defer close(ch)

		ctx, cancel := context.WithCancel(context.Background())

		time.AfterFunc(time.Millisecond*10, func() {
			cancel()
		})
		err := prompt(ctx, blockingReader{ch})
		assert.EqualError(t, err, "prompt canceled")
	})

	t.Run("user cancel", func(t *testing.T) {
		err := prompt(context.Background(), strings.NewReader("n\n"))
		assert.EqualError(t, err, "user refused execution")

		err = prompt(context.Background(), strings.NewReader("N\n"))
		assert.EqualError(t, err, "user refused execution")

		err = prompt(context.Background(), strings.NewReader("J\n"))
		assert.EqualError(t, err, "user refused execution")
	})

	t.Run("user accept", func(t *testing.T) {
		assert.NoError(t, prompt(context.TODO(), strings.NewReader("y\n")))
		assert.NoError(t, prompt(context.TODO(), strings.NewReader("Y\n")))
		assert.NoError(t, prompt(context.TODO(), strings.NewReader("\n")))
	})

	t.Run("faulty reader", func(t *testing.T) {
		assert.Error(t, prompt(context.TODO(), iotest.ErrReader(errors.New("phony error"))))
	})
}

type blockingReader struct{ close <-chan struct{} }

func (b blockingReader) Read(p []byte) (n int, err error) {
	<-b.close
	return 0, errors.New("reader closed")

}

func Test_replaceArgs(t *testing.T) {
	t.Run("no replacement requested", func(t *testing.T) {
		v, err := replaceArgs(nil, "")("ctx")
		assert.NoError(t, err)
		assert.Equal(t, []string{"--context=ctx"}, v)

		v, err = replaceArgs([]string{"arg1", "arg2"}, "")("ctx")
		assert.NoError(t, err)
		assert.Equal(t, []string{"--context=ctx", "arg1", "arg2"}, v)
	})
	t.Run("no replacement happened", func(t *testing.T) {
		_, err := replaceArgs([]string{}, "X")("ctx")
		assert.Error(t, err)
		_, err = replaceArgs([]string{"--arg1", "--arg2"}, "X")("ctx")
		assert.Error(t, err)
	})
	t.Run("replacement hits", func(t *testing.T) {
		v, err := replaceArgs([]string{"a", "XX", "aX"}, "X")("ctx")
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "ctxctx", "actx"}, v)

		v, err = replaceArgs([]string{"a", "XX", "aX"}, "XX")("ctx")
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "ctx", "aX"}, v)
	})
}

func Test_findKubeContextArg(t *testing.T) {
	t.Run("no context", func(t *testing.T) {
		v, ok := findKubeContextArg([]string{"--arg1", "--arg2"})
		assert.False(t, ok)
		assert.Empty(t, v)
	})
	t.Run("context", func(t *testing.T) {
		v, ok := findKubeContextArg([]string{"--context=ctx", "--arg1", "--arg2"})
		assert.True(t, ok)
		assert.Equal(t, "--context=ctx", v)
	})
	t.Run("context", func(t *testing.T) {
		v, ok := findKubeContextArg([]string{"--context", "ctx", "--arg1", "--arg2"})
		assert.True(t, ok)
		assert.Equal(t, "--context", v)
	})
}

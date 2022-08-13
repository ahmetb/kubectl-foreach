package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
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
}

type blockingReader struct{ close <-chan struct{} }

func (b blockingReader) Read(p []byte) (n int, err error) {
	<-b.close
	return 0, fmt.Errorf("reader closed")

}

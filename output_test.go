package main

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_synchronizedWriter(t *testing.T) {
	var b bytes.Buffer
	sw := &synchronizedWriter{Writer: &b}
	n := 1000
	var wg sync.WaitGroup
	wg.Add(n)
	seq := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < n; i++ {
		go func() {
			sw.Write([]byte(seq))
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1000, strings.Count(b.String(), seq))
}

func Test_prefixingWriter(t *testing.T) {
	var b bytes.Buffer
	pw := &prefixingWriter{prefix: []byte{'p', ':', ' '}, w: &b}

	// single line (no trailing newline)
	n, err := pw.Write([]byte("hello"))
	assert.Equal(t, 5, n)
	assert.NoError(t, err)

	// multi line (no trailing newline) - continuation to single line + new line
	n, err = pw.Write([]byte("a\nb"))
	assert.Equal(t, 3, n)
	assert.NoError(t, err)
	assert.Equal(t, `p: helloa
`, b.String())

	n, err = pw.Write([]byte("eof\n"))
	assert.Equal(t, 4, n)
	assert.NoError(t, err)

	assert.Equal(t, `p: helloa
p: beof
`, // expected trailing newline
		b.String())
}

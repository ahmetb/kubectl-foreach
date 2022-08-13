package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFilter(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    filter
		wantErr require.ErrorAssertionFunc
	}{
		{name: "empty spec",
			in:      "",
			wantErr: require.Error},
		{name: "exact match",
			in:      "foo",
			want:    exact("foo"),
			wantErr: require.NoError},
		{name: "exact match inverted",
			in:      "^foo",
			want:    exclude{exact("foo")},
			wantErr: require.NoError},
		{name: "pattern",
			in:      "/re/",
			want:    pattern{regexp.MustCompile("re")},
			wantErr: require.NoError},
		{name: "pattern missing trailing slash",
			in:      "/re",
			want:    exact("/re"),
			wantErr: require.NoError},
		{name: "pattern parse error",
			in:      "/re(/",
			wantErr: require.Error},
		{name: "pattern inverted",
			in:      "^/re/",
			want:    exclude{pattern{regexp.MustCompile("re")}},
			wantErr: require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilter(tt.in)
			tt.wantErr(t, err, fmt.Sprintf("parseFilter(%q)", tt.in))
			assert.Equalf(t, tt.want, got, "parseFilter(%q)", tt.in)
		})
	}
}

func TestExact(t *testing.T) {
	v := exact("foo")
	assert.True(t, v.additive())
	assert.True(t, v.match("foo"))
	assert.False(t, v.match("bar"))
}

func TestPattern(t *testing.T) {
	v := pattern{regexp.MustCompile("^re")}
	assert.True(t, v.additive())
	assert.False(t, v.match("are"))
	assert.True(t, v.match("res"))
}

func TestExclude(t *testing.T) {
	v := exclude{exact("foo")}
	assert.False(t, v.additive())
	assert.False(t, v.match("bar"))
	assert.True(t, v.match("foo"))
}

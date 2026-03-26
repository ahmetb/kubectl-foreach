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
	t.Run("uses the leftmost double dash", func(t *testing.T) {
		l, r, err := separateArgs([]string{"a", "b", "--", "foo", "--", "--bar"})
		assert.Nil(t, err)
		assert.Equal(t, []string{"a", "b"}, l)
		assert.Equal(t, []string{"foo", "--", "--bar"}, r)
	})
}

func Test_detectOutputFormat(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"empty args", nil, ""},
		{"no output flag", []string{"get", "pods"}, ""},
		{"-ojson", []string{"get", "pods", "-ojson"}, "json"},
		{"-oyaml", []string{"get", "pods", "-oyaml"}, "yaml"},
		{"-o=json", []string{"get", "pods", "-o=json"}, "json"},
		{"-o=yaml", []string{"get", "pods", "-o=yaml"}, "yaml"},
		{"-o json", []string{"get", "pods", "-o", "json"}, "json"},
		{"-o yaml", []string{"get", "pods", "-o", "yaml"}, "yaml"},
		{"--output=json", []string{"get", "pods", "--output=json"}, "json"},
		{"--output=yaml", []string{"get", "pods", "--output=yaml"}, "yaml"},
		{"--output json", []string{"get", "pods", "--output", "json"}, "json"},
		{"--output yaml", []string{"get", "pods", "--output", "yaml"}, "yaml"},
		{"other format ignored", []string{"get", "pods", "-o", "wide"}, ""},
		{"jsonpath ignored", []string{"get", "pods", "-o=jsonpath={.items}"}, ""},
		{"-o at end without value", []string{"get", "pods", "-o"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, detectOutputFormat(tt.args))
		})
	}
}

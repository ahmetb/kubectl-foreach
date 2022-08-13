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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_matchContexts(t *testing.T) {
	type args struct {
		in []string
		f  []filter
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "empty input",
			args: args{
				in: nil,
				f:  []filter{exact("foo")}},
			want: nil},
		{name: "empty filters match all",
			args: args{
				in: []string{"a", "b", "c"},
				f:  []filter{}},
			want: []string{"a", "b", "c"}},
		{name: "only additive patterns",
			args: args{
				in: []string{"a", "b", "c"},
				f:  []filter{exact("a"), pattern{regexp.MustCompile("^c")}}},
			want: []string{"a", "c"}},
		{name: "only additive patterns no results",
			args: args{
				in: []string{"a", "b", "c"},
				f:  []filter{exact("d"), pattern{regexp.MustCompile("^e")}}},
			want: nil},
		{name: "only excluding patterns",
			args: args{
				in: []string{"a", "b", "c"},
				f:  []filter{exclude{exact("b")}, exclude{exact("d")}}},
			want: []string{"a", "c"}},
		{name: "only excluding patterns no results",
			args: args{
				in: []string{"a", "b", "c"},
				f:  []filter{exclude{pattern{regexp.MustCompile(`^`)}}}},
			want: nil},
		{name: "mixed patterns",
			args: args{
				in: []string{"a", "b", "c", "d", "e"},
				f: []filter{
					exact("a"),
					exact("b"),
					exclude{exact("b")},
					exclude{exact("e")},
					pattern{regexp.MustCompile("^[cde]")},
				}},
			want: []string{"a", "c", "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, matchContexts(tt.args.in, tt.args.f), "matchContexts(%v, %v)", tt.args.in, tt.args.f)
		})
	}
}

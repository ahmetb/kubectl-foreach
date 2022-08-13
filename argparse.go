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
	"fmt"
)

// separateArgs parses command-line arguments (excluding argv[0]) meant for the tool and kubectl
// (separated by '--', which is removed during separation).
func separateArgs(argv []string) (toolArgs []string, kubectlArgs []string, err error) {
	var separatorFound bool
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			if i == len(argv)-1 {
				err = fmt.Errorf("need to specify arguments to kubectl after '--'")
				return
			}
			separatorFound = true
			kubectlArgs = argv[i+1:]
			break
		}
		toolArgs = append(toolArgs, arg)
	}
	if !separatorFound {
		err = fmt.Errorf("need to specify the '--' as an argument, followed by arguments to kubectl")
		return
	}
	return
}

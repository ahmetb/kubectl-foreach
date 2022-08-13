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

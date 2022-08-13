package main

import (
	"fmt"
)

// parseArgs parses positional arguments (left after processing options) into context matchers and subcommand
// arguments (split by '--').
func parseArgs(args []string) (filters []filter, cmdArgs []string, err error) {
	var separatorFound bool
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i == len(args)-1 {
				err = fmt.Errorf("need to specify arguments to kubectl after '--'")
				return
			}
			separatorFound = true
			cmdArgs = args[i+1:]
			break
		}

		f, ferr := parseFilter(arg)
		if ferr != nil {
			err = ferr
			return
		}
		filters = append(filters, f)
	}
	if !separatorFound {
		err = fmt.Errorf("need to specify the '--' as an argument, followed by arguments to kubectl")
		return
	}
	return
}

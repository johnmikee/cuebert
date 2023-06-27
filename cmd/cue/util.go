package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func extractFlagValue(fn string, args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Check if the argument matches the flag in any of the supported formats
		if arg == "--"+fn || arg == "-"+fn {
			// If the next argument is not another flag, return it as the flag value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				return args[i+1], nil
			}
			return "", fmt.Errorf("flag '%s' requires a value", fn)
		} else if strings.HasPrefix(arg, "--"+fn+"=") || strings.HasPrefix(arg, "-"+fn+"=") {
			// If the flag is in the format "--flag=value" or "-flag=value", extract the value
			return strings.SplitN(arg, "=", 2)[1], nil
		}
	}

	return "", fmt.Errorf("flag '%s' not found", fn)
}

func validateAddArgs(args []string) bool {
	for _, i := range args {
		if i == "" {
			return false
		}
	}

	return true
}

func shellCMD(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.Output()
	if err != nil {
		return string(stdout), err
	}

	return string(stdout), nil
}

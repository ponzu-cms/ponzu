package main

import "runtime"

// buildOutputName returns the correct ponzu-server file name
// based on the host Operating System
func buildOutputName() string {
	if runtime.GOOS == "windows" {
		return "ponzu-server.exe"
	}

	return "ponzu-server"
}

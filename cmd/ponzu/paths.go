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

// buildOutputPath returns the correct path to the ponzu-server binary
// built, based on the host Operating System. This is necessary so that
// the UNIX-y systems know to look in the current directory, and not the $PATH
func buildOutputPath() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return "."
}

package main

import "runtime"

// GetPonzuServerBuildOutputFileName returns de correct
// ponzu-server file name based on the host Operating System
func GetPonzuServerBuildOutputFileName() string {
	if runtime.GOOS == "windows" {
		return "ponzu-server.exe"
	}
	return "ponzu-server"
}

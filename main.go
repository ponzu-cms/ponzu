// Package main is located in the cmd/ponzu directory and contains the code to build
// and operate the command line interface (CLI) to manage Ponzu systems. Here,
// you will find the code that is used to create new Ponzu projects, generate
// code for content types and other uploads, build Ponzu binaries and run servers.
package main

import (
	_ "github.com/fanky5g/ponzu/internal/domain/entities/content"
	"github.com/fanky5g/ponzu/internal/handler/command"
)

func main() {
	command.Execute()
}

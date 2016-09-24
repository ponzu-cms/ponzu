package main

import (
	"flag"
	"fmt"
	"os"
)

var usage = `
$ cms <option> <params>

Options: 
generate, gen, g:
    Generate a new content type file with boilerplate code to implement
    the editor.Editable interface. Must be given one (1) parameter of
    the name of the type for the new content.

    Example:
        $ cms gen Review

`

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	fmt.Println(args)
	switch args[0] {
	case "generate", "gen", "g":
		if len(args) < 2 {
			flag.PrintDefaults()
			os.Exit(0)
		}

		name := args[1]

		err := generateContentType(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "serve", "s":
		serve()
	default:
		flag.PrintDefaults()
	}
}

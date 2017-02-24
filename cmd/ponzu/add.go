package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Use `go get` to download addon and add to $GOPATH/src, useful
// for IDE auto-import and code completion, then copy entire directory
// tree to project's ./addons folder
func getAddon(args []string) error {

	var cmdOptions []string
	var addonPath = args[1]

	// Go get
	cmdOptions = append(cmdOptions, "get", addonPath)
	get := exec.Command(gocmd, cmdOptions...)
	get.Stderr = os.Stderr
	get.Stdout = os.Stdout

	err := get.Start()
	if err != nil {
		addError(err)
	}
	err = get.Wait()
	if err != nil {
		addError(err)
	}

	// Copy to ./addons folder
	// GOPATH can be a list delimited by ":" on Linux or ";" on Windows
	// `go get` uses the first, this should parse out the first whatever the OS
	gopath := resolveGOPATH()

	pwd, err := os.Getwd()
	if err != nil {
		addError(err)
	}

	src := filepath.Join(gopath, "src", addonPath)
	dest := filepath.Join(pwd, "addons", addonPath)
	log.Println(dest)

	err = os.Mkdir(dest, os.ModeDir|os.ModePerm)
	if err != nil {
		addError(err)
	}
	err = copyAll(src, dest)
	if err != nil {
		log.Println(err)
		//addError(err)
	}
	return nil
}

// GOPATH can be a list delimited by ":" on Linux or ";" on Windows
// `go get` uses saves packages to the first entry, so this function
// should parse out the first whatever the OS
func resolveGOPATH() string {
	envGOPATH := os.Getenv("GOPATH")
	gopaths := strings.Split(envGOPATH, ":")
	gopath := gopaths[0]
	gopaths = strings.Split(envGOPATH, ";")
	gopath = gopaths[0]
	return gopath
}

// error return
func addError(err error) error {
	return errors.New("Ponzu add failed. " + "\n" + err.Error())
}

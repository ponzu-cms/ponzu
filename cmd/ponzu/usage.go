package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var year = fmt.Sprintf("%d", time.Now().Year())

var usageHeader = `
$ ponzu [flags] command <params>

Ponzu is a powerful and efficient open-source HTTP server framework and CMS. It 
provides automatic, free, and secure HTTP/2 over TLS (certificates obtained via 
[Let's Encrypt](https://letsencrypt.org)), a useful CMS and scaffolding to 
generate set-up code, and a fast HTTP API on which to build modern applications.

Ponzu is released under the BSD-3-Clause license (see LICENSE).
(c) 2016 - ` + year + ` Boss Sauce Creative, LLC

COMMANDS:

`

var usageHelp = `
help, h (command)

	Help command will print the usage for Ponzu, or if a command is entered, it
	will show only the usage for that specific command.

	Example:
	$ ponzu help generate
	
	
`

var usageNew = `
new <directory>

	Creates a 'ponzu' directory, or one by the name supplied as a parameter 
	immediately following the 'new' option in the $GOPATH/src directory. Note: 
	'new' depends on the program 'git' and possibly a network connection. If 
	there is no local repository to clone from at the local machine's $GOPATH, 
	'new' will attempt to clone the 'github.com/ponzu-cms/ponzu' package from 
	over the network.

	Example:
	$ ponzu new myProject
	> New ponzu project created at $GOPATH/src/myProject

	Errors will be reported, but successful commands retrun nothing.


`

var usageGenerate = `
generate, gen, g <generator type (,...fields)>

	Generate boilerplate code for various Ponzu components, such as 'content'.

	Example:
	$ ponzu gen content review title:"string" body:"string" rating:"int" tags:"[]string"

	The command above will generate a file 'content/review.go' with boilerplate
	methods, as well as struct definition, and cooresponding field tags like:

	type Review struct {
		Title  string   ` + "`json:" + `"title"` + "`" + `
		Body   string   ` + "`json:" + `"body"` + "`" + `
		Rating int      ` + "`json:" + `"rating"` + "`" + `
		Tags   []string ` + "`json:" + `"tags"` + "`" + `
	}

	The generate command will intelligently parse more sophisticated field names
	such as 'field_name' and convert it to 'FieldName' and vice versa, only where 
	appropriate as per common Go idioms. Errors will be reported, but successful 
	generate commands retrun nothing.


`

var usageBuild = `
[-gocmd=go] build

	From within your Ponzu project directory, running build will copy and move 
	the necessary files from your workspace into the vendored directory, and 
	will build/compile the project to then be run. 
	
	Example:
	$ ponzu build
	(or)
	$ ponzu -gocmd=go1.8rc1 build

	By providing the 'gocmd' flag, you can specify which Go command to build the
	project, if testing a different release of Go.

	Errors will be reported, but successful build commands return nothing.


`

var usageRun = `
[[-port=8080] [--https|--devhttps]] run <service(,service)>

	Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
	The segments, separated by a comma, describe which services to start, either 
	'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally, 
	if the server should utilize TLS encryption - served over HTTPS, which is
	automatically managed using Let's Encrypt (https://letsencrypt.org) 

	Example: 
	$ ponzu run
	(or)
	$ ponzu -port=8080 --https run admin,api
	(or) 
	$ ponzu run admin
	(or)
	$ ponzu -port=8888 run api

	Defaults to '-port=8080 run admin,api' (running Admin & API on port 8080, without TLS)

	Note: 
	Admin and API cannot run on separate processes unless you use a copy of the
	database, since the first process to open it receives a lock. If you intend
	to run the Admin and API on separate processes, you must call them with the
	'ponzu' command independently.


`

var usageUpgrade = `
upgrade

	Will backup your own custom project code (like content, addons, uploads, etc) so
	we can safely re-clone Ponzu from the latest version you have or from the network 
	if necessary. Before running '$ ponzu upgrade', you should update the 'ponzu'
	package by running '$ go get -u github.com/ponzu-cms/ponzu/...' 

	Example:
	$ ponzu upgrade


`

var usageVersion = `
[--cli] version, v

	Prints the version of Ponzu your project is using. Must be called from 
	within a Ponzu project directory.

	Example:
	$ ponzu version
	> Ponzu v0.7.1
	(or)
	$ ponzu --cli version
	> Ponzu v0.7.2


`

func ponzu(isCLI bool) (map[string]interface{}, error) {
	kv := make(map[string]interface{})

	info := filepath.Join("cmd", "ponzu", "ponzu.json")
	if isCLI {
		gopath := os.Getenv("GOPATH")
		repo := filepath.Join(gopath, "src", "github.com", "ponzu-cms", "ponzu")
		info = filepath.Join(repo, "cmd", "ponzu", "ponzu.json")
	}

	b, err := ioutil.ReadFile(info)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &kv)
	if err != nil {
		return nil, err
	}

	return kv, nil
}

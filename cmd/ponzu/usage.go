package main

import (
	"fmt"
	"time"
)

var year = fmt.Sprintf("%d", time.Now().Year())

var usageHeader = `
$ ponzu [flags] command <params>

Ponzu is a powerful and efficient open-source "Content-as-a-Service" system 
framework and CMS. It provides automatic, free, and secure HTTP/2 over TLS 
(certificates obtained via Let's Encrypt - https://letsencrypt.org), a useful 
CMS and  scaffolding to generate content editors, and a fast HTTP API on which 
to build modern applications.

Ponzu is released under the BSD-3-Clause license (see LICENSE).
(c) ` + year + ` Boss Sauce Creative, LLC

COMMANDS

`

var usageHelp = `
help, h (command):

	Help command will print the usage for Ponzu, or if a command is entered, it
	will show only the usage for that specific command.

	Example:
	$ ponzu help generate
	
	
`

var usageNew = `
new <directory>:

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
generate, gen, g <type (,...fields)>:

	Generate a content type file with boilerplate code to implement
	the editor.Editable interface. Must be given one (1) parameter of
	the name of the type for the new content. The fields following a 
	type determine the field names and types of the content struct to 
	be generated. These must be in the following format:
	fieldName:"T"

	Example:
	$ ponzu gen review title:"string" body:"string" rating:"int" tags:"[]string"

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
build

	From within your Ponzu project directory, running build will copy and move 
	the necessary files from your workspace into the vendored directory, and 
	will build/compile the project to then be run. 
	
	Example:
	$ ponzu build

	Errors will be reported, but successful build commands return nothing.

`

var usageRun = `
[[--port=8080] [--https]] run <service(,service)>:

	Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
	The segments, separated by a comma, describe which services to start, either 
	'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally, 
	if the server should utilize TLS encryption - served over HTTPS, which is
	automatically managed using Let's Encrypt (https://letsencrypt.org) 

	Example: 
	$ ponzu run
	(or)
	$ ponzu --port=8080 --https run admin,api
	(or) 
	$ ponzu run admin
	(or)
	$ ponzu --port=8888 run api

	Defaults to '--port=8080 run admin,api' (running Admin & API on port 8080, without TLS)

	Note: 
	Admin and API cannot run on separate processes unless you use a copy of the
	database, since the first process to open it receives a lock. If you intend
	to run the Admin and API on separate processes, you must call them with the
	'ponzu' command independently.


`

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bosssauce/ponzu/system/admin"
	"github.com/bosssauce/ponzu/system/api"
	"github.com/bosssauce/ponzu/system/api/analytics"
	"github.com/bosssauce/ponzu/system/db"
	"github.com/bosssauce/ponzu/system/tls"

	// import registers content types
	_ "github.com/bosssauce/ponzu/content"
)

var year = fmt.Sprintf("%d", time.Now().Year())

var usageHeader = `
$ ponzu [specifiers] command <params>

Ponzu is a powerful and efficient open-source "Content-as-a-Service" system 
framework. It provides automatic, free, and secure HTTP/2 over TLS (certificates 
obtained via Let's Encrypt - https://letsencrypt.org), a useful CMS and 
scaffolding to generate content editors, and a fast HTTP API on which to build 
modern applications.

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

	Creates a 'ponzu' directorty, or one by the name supplied as a parameter 
	immediately following the 'new' option in the $GOPATH/src directory. Note: 
	'new' depends on the program 'git' and possibly a network connection. If 
	there is no local repository to clone from at the local machine's $GOPATH, 
	'new' will attempt to clone the 'github.com/bosssauce/ponzu' package from 
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
	database, since the first process to open it recieves a lock. If you intend
	to run the Admin and API on separate processes, you must call them with the
	'ponzu' command independently.


`

var (
	usage = usageHeader + usageNew + usageGenerate + usageBuild + usageRun
	port  int
	https bool

	// for ponzu internal / core development
	dev  bool
	fork string
)

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
}

func main() {
	flag.IntVar(&port, "port", 8080, "port for ponzu to bind its listener")
	flag.BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
	flag.BoolVar(&dev, "dev", false, "modify environment for Ponzu core development")
	flag.StringVar(&fork, "fork", "", "modify repo source for Ponzu core development")
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Println(usage)
		os.Exit(0)
	}

	switch args[0] {
	case "help", "h":
		if len(args) < 2 {
			fmt.Println(usageHelp)
			fmt.Println(usage)
			os.Exit(0)
		}

		switch args[1] {
		case "new":
			fmt.Println(usageNew)
			os.Exit(0)

		case "generate", "gen", "g":
			fmt.Println(usageGenerate)
			os.Exit(0)

		case "build":
			fmt.Println(usageBuild)
			os.Exit(0)

		case "run":
			fmt.Println(usageRun)
			os.Exit(0)
		}

	case "new":
		if len(args) < 2 {
			fmt.Println(usage)
			os.Exit(0)
		}

		err := newProjectInDir(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "generate", "gen", "g":
		if len(args) < 2 {
			flag.PrintDefaults()
			os.Exit(0)
		}

		err := generateContentType(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "build":
		err := buildPonzuServer(args)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "run":
		var addTLS string
		if https {
			addTLS = "--https"
		} else {
			addTLS = "--https=false"
		}

		var services string
		if len(args) > 1 {
			services = args[1]
		} else {
			services = "admin,api"
		}

		serve := exec.Command("./ponzu-server",
			fmt.Sprintf("--port=%d", port),
			addTLS,
			"serve",
			services,
		)
		serve.Stderr = os.Stderr
		serve.Stdout = os.Stdout

		err := serve.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = serve.Wait()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "serve", "s":
		db.Init()
		defer db.Close()

		analytics.Init()
		defer analytics.Close()

		if len(args) > 1 {
			services := strings.Split(args[1], ",")

			for i := range services {
				if services[i] == "api" {
					api.Run()
				} else if services[i] == "admin" {
					admin.Run()
				} else {
					fmt.Println("To execute 'ponzu serve', you must specify which service to run.")
					fmt.Println("$ ponzu --help")
					os.Exit(1)
				}
			}
		}

		if https {
			fmt.Println("Enabling HTTPS...")
			tls.Enable()
		}

		// save the port the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		fmt.Println(port, "port from main")
		err := db.PutConfig("http_port", fmt.Sprintf("%d", port))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.")
		}

		cfg, _ := db.ConfigAll()
		fmt.Println(string(cfg))

		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

	case "":
		fmt.Println(usage)
		fmt.Println(usageHelp)

	default:
		fmt.Println(usage)
		fmt.Println(usageHelp)
	}
}

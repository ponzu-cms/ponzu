package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bosssauce/ponzu/system/admin"
	"github.com/bosssauce/ponzu/system/api"
	"github.com/bosssauce/ponzu/system/api/analytics"
	"github.com/bosssauce/ponzu/system/db"
	"github.com/bosssauce/ponzu/system/tls"
)

var usage = `
$ ponzu [specifiers] option <params>

Options 

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



generate, gen, g <type>:

	Generate a content type file with boilerplate code to implement
	the editor.Editable interface. Must be given one (1) parameter of
	the name of the type for the new content.

	Example:
	$ ponzu gen review



[[--port=8080] [--https]] run <service(,service)>:

	Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
	The segments, separated by a comma, describe which services to start, either 
	'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally, 
	if the server(s) should utilize TLS encryption (served over HTTPS), which is
	automatically managed using Let's Encrypt (https://letsencrypt.org) 

	Example: 
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
		flag.Usage()
		os.Exit(0)
	}

	switch args[0] {
	case "new":
		if len(args) < 2 {
			flag.PrintDefaults()
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

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

	case "":
		flag.PrintDefaults()

	default:
		flag.PrintDefaults()
	}
}

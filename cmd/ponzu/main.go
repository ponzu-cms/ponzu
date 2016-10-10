package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bosssauce/ponzu/system/admin"
	"github.com/bosssauce/ponzu/system/api"
	"github.com/bosssauce/ponzu/system/db"
)

var usage = `
$ ponzu option <params> [specifiers]

Options 

new <directory>:

	Creates a new 'ponzu' in the current directory, or one supplied
	as a parameter immediately following the 'new' option. Note: 'new'
	depends on the program 'git' and possibly a network connection. If there is
	no local repository to clone from at the local machine's $GOPATH, 'new' will
	attempt to clone the 'ponzu' package from over the network.

	Example:
	$ ponzu new ~/Projects/my-project.dev



generate, gen, g <type>:

    Generate a content type file with boilerplate code to implement
    the editor.Editable interface. Must be given one (1) parameter of
    the name of the type for the new content.

    Example:
	$ ponzu gen review



[[--port=8080] [--tls]] serve, s <service(,service)>:

	Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
	Must be given at least one (1) parameter. The segments describe 
	which services to start, either 'admin' (Admin System / CMS 
	backend) or 'api' (JSON API), and, optionally, if the server(s) should 
	utilize TLS encryption (served over HTTPS), which is automatically managed 
	using Let's Encrypt (https://letsencrypt.org) 

	Example: 
	$ ponzu --port=8080 --tls serve admin,api
	(or) 
	$ ponzu serve admin
	(or)
	$ ponzu --port=8888 serve api

	Defaults to '--port=8080 admin,api' (running Admin & API on port 8080, without TLS)

	Note: 
	Admin and API cannot run on separate processes unless you use a copy of the
	database, since the first process to open it recieves a lock. If you intend
	to run the Admin and API on separate processes, you must call them with the
	'ponzu' command independently.
`

var (
	port int
	tls  bool
)

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
	}
}

func main() {
	flag.IntVar(&port, "port", 8080, "port for ponzu to bind its listener")
	flag.BoolVar(&tls, "tls", false, "enable automatic TLS/SSL certificate management")
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

		err := generateContentType(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "serve", "s":
		db.Init()
		if len(args) > 1 {
			services := strings.Split(args[1], ",")

			fmt.Println(args, port, tls)
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
		} else {
			if len(args) > 1 {
				if args[1] == "admin" {
					admin.Run()
				}

				if args[1] == "api" {
					api.Run()
				}
			} else {
				admin.Run()
				api.Run()
			}

		}

		if tls {
			fmt.Println("TLS through Let's Encrypt is not implemented yet.")
			fmt.Println("Please run 'ponzu serve' without the -tls flag for now.")
			os.Exit(1)
		}

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	case "":
		flag.PrintDefaults()
	default:
		flag.PrintDefaults()
	}
}

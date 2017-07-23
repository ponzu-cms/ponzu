// Package main is located in the cmd/ponzu directory and contains the code to build
// and operate the command line interface (CLI) to manage Ponzu systems. Here,
// you will find the code that is used to create new Ponzu projects, generate
// code for content types and other files, build Ponzu binaries and run servers.
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/ponzu-cms/ponzu/content"
	"github.com/ponzu-cms/ponzu/system/admin"
	"github.com/ponzu-cms/ponzu/system/api"
	"github.com/ponzu-cms/ponzu/system/api/analytics"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/tls"

	"github.com/spf13/cobra"
)

var (
	bind      string
	httpsport int
	port      int
	docsport  int
	https     bool
	devhttps  bool
	docs      bool
	cli       bool

	// for ponzu internal / core development
	gocmd string
	fork  string
	dev   bool

	year = fmt.Sprintf("%d", time.Now().Year())
)

var rootCmd = &cobra.Command{
	Use: "ponzu",
	Long: `Ponzu is an open-source HTTP server framework and CMS, released under 
the BSD-3-Clause license.
(c) 2016 - ` + year + ` Boss Sauce Creative, LLC`,
}

var runCmd = &cobra.Command{
	Use:   "run [flags] <service(,service)>",
	Short: "starts the 'ponzu' HTTP server for the JSON API and or Admin System.",
	Long: `Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
The segments, separated by a comma, describe which services to start, either
'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally,
if the server should utilize TLS encryption - served over HTTPS, which is
automatically managed using Let's Encrypt (https://letsencrypt.org)

Defaults to 'run --port=8080 admin,api' (running Admin & API on port 8080, without TLS)

Note:
Admin and API cannot run on separate processes unless you use a copy of the
database, since the first process to open it receives a lock. If you intend
to run the Admin and API on separate processes, you must call them with the
'ponzu' command independently.`,
	Example: `$ ponzu run
(or)
$ ponzu run --port=8080 --https admin,api
(or)
$ ponzu run admin
(or)
$ ponzu run --port=8888 api`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var addTLS string
		if https {
			addTLS = "--https"
		} else {
			addTLS = "--https=false"
		}

		if devhttps {
			addTLS = "--dev-https"
		}

		var addDocs string
		if docs {
			addDocs = "--docs"
		} else {
			addDocs = "--docs=false"
		}

		var services string
		if len(args) > 0 {
			services = args[0]
		} else {
			services = "admin,api"
		}

		name := buildOutputName()
		buildPathName := strings.Join([]string{".", name}, string(filepath.Separator))
		serve := exec.Command(buildPathName,
			"serve",
			services,
			fmt.Sprintf("--bind=%s", bind),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--https-port=%d", httpsport),
			fmt.Sprintf("--docs-port=%d", docsport),
			addDocs,
			addTLS,
		)
		serve.Stderr = os.Stderr
		serve.Stdout = os.Stdout

		return serve.Run()
	},
}

// ErrWrongOrMissingService informs a user that the services to run must be
// explicitly specified when serve is called
var ErrWrongOrMissingService = errors.New("To execute 'ponzu serve', " +
	"you must specify which service to run.")

var serveCmd = &cobra.Command{
	Use:     "serve [flags] <service,service>",
	Aliases: []string{"s"},
	Short:   "run the server (serve is wrapped by the run command)",
	Hidden:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrWrongOrMissingService
		}

		db.Init()
		defer db.Close()

		analytics.Init()
		defer analytics.Close()

		services := strings.Split(args[0], ",")

		for _, service := range services {
			if service == "api" {
				api.Run()
			} else if service == "admin" {
				admin.Run()
			} else {
				return ErrWrongOrMissingService
			}
		}

		// run docs server if --docs is true
		if docs {
			admin.Docs(docsport)
		}

		// init search index
		go db.InitSearchIndex()

		// save the https port the system is listening on
		err := db.PutConfig("https_port", fmt.Sprintf("%d", httpsport))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// cannot run production HTTPS and development HTTPS together
		if devhttps {
			fmt.Println("Enabling self-signed HTTPS... [DEV]")

			go tls.EnableDev()
			fmt.Println("Server listening on https://localhost:10443 for requests... [DEV]")
			fmt.Println("----")
			fmt.Println("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
			fmt.Println("on Chrome, visit chrome://flags/#allow-insecure-localhost")

		} else if https {
			fmt.Println("Enabling HTTPS...")

			go tls.Enable()
			fmt.Printf("Server listening on :%s for HTTPS requests...\n", db.ConfigCache("https_port").(string))
		}

		// save the https port the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		err = db.PutConfig("http_port", fmt.Sprintf("%d", port))
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		// save the bound address the system is listening on so internal system can make
		// HTTP api calls while in dev or production w/o adding more cli flags
		if bind == "" {
			bind = "localhost"
		}
		err = db.PutConfig("bind_addr", bind)
		if err != nil {
			log.Fatalln("System failed to save config. Please try to run again.", err)
		}

		fmt.Printf("Server listening at %s:%d for HTTP requests...\n", bind, port)
		fmt.Println("\nVisit '/admin' to get started.")
		log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%d", bind, port), nil))
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{runCmd, serveCmd} {
		cmd.Flags().StringVar(&bind, "bind", "localhost", "address for ponzu to bind the HTTP(S) server")
		cmd.Flags().IntVar(&httpsport, "https-port", 443, "port for ponzu to bind its HTTPS listener")
		cmd.Flags().IntVar(&port, "port", 8080, "port for ponzu to bind its HTTP listener")
		cmd.Flags().IntVar(&docsport, "docs-port", 1234, "[dev environment] override the documentation server port")
		cmd.Flags().BoolVar(&docs, "docs", false, "[dev environment] run HTTP server to view local HTML documentation")
		cmd.Flags().BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
		cmd.Flags().BoolVar(&devhttps, "dev-https", false, "[dev environment] enable automatic TLS/SSL certificate management")
	}

	RegisterCmdlineCommand(serveCmd)
	RegisterCmdlineCommand(runCmd)

	pflags := rootCmd.PersistentFlags()
	pflags.StringVar(&gocmd, "gocmd", "go", "custom go command if using beta or new release of Go")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execAndWait(command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return err

	}
	return cmd.Wait()
}

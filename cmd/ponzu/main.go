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

	"github.com/ponzu-cms/ponzu/system/admin"
	"github.com/ponzu-cms/ponzu/system/api"
	"github.com/ponzu-cms/ponzu/system/api/analytics"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/tls"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/ponzu-cms/ponzu/content"
)

var (
	port      int
	httpsport int
	https     bool
	devhttps  bool
	cli       bool

	// for ponzu internal / core development
	dev   bool
	fork  string
	gocmd string
	year  = fmt.Sprintf("%d", time.Now().Year())
)

var rootCmd = &cobra.Command{
	Use: "ponzu",
	Long: `Ponzu is a powerful and efficient open-source HTTP server framework and CMS. It
provides automatic, free, and secure HTTP/2 over TLS (certificates obtained via
[Let's Encrypt](https://letsencrypt.org)), a useful CMS and scaffolding to
generate set-up code, and a fast HTTP API on which to build modern applications.

Ponzu is released under the BSD-3-Clause license (see LICENSE).
(c) 2016 - ` + year + ` Boss Sauce Creative, LLC`,
}

var runCmd = &cobra.Command{
	Use:   "run <service(,service)>",
	Short: "starts the 'ponzu' HTTP server for the JSON API and or Admin System.",
	Long: `Starts the 'ponzu' HTTP server for the JSON API, Admin System, or both.
The segments, separated by a comma, describe which services to start, either
'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally,
if the server should utilize TLS encryption - served over HTTPS, which is
automatically managed using Let's Encrypt (https://letsencrypt.org)

Defaults to '-port=8080 run admin,api' (running Admin & API on port 8080, without TLS)

Note:
Admin and API cannot run on separate processes unless you use a copy of the
database, since the first process to open it receives a lock. If you intend
to run the Admin and API on separate processes, you must call them with the
'ponzu' command independently.`,
	Example: `$ ponzu run
(or)
$ ponzu -port=8080 --https run admin,api
(or)
$ ponzu run admin
(or)
$ ponzu -port=8888 run api`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var addTLS string
		if https {
			addTLS = "--https"
		} else {
			addTLS = "--https=false"
		}

		if devhttps {
			addTLS = "--devhttps"
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
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--httpsport=%d", httpsport),
			addTLS,
			"serve",
			services,
		)
		serve.Stderr = os.Stderr
		serve.Stdout = os.Stdout

		return serve.Run()
	},
}

var ErrWrongOrMissingService = errors.New("To execute 'ponzu serve', " +
	"you must specify which service to run.")

var serveCmd = &cobra.Command{
	Use:     "serve <service,service>",
	Aliases: []string{"s"},
	Short:   "actually run the server",
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

		fmt.Printf("Server listening on :%d for HTTP requests...\n", port)
		fmt.Println("\nvisit `/admin` to get started.")
		log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{runCmd, serveCmd} {
		cmd.Flags().IntVar(&port, "port", 8080, "port for ponzu to bind its HTTP listener")
		cmd.Flags().IntVar(&httpsport, "httpsport", 443, "port for ponzu to bind its HTTPS listener")
		cmd.Flags().BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
		cmd.Flags().BoolVar(&devhttps, "devhttps", false, "[dev environment] enable automatic TLS/SSL certificate management")
	}

	rootCmd.AddCommand(runCmd, serveCmd)

	pflags := rootCmd.PersistentFlags()
	pflags.StringVar(&gocmd, "gocmd", "go", "custom go command if using beta or new release of Go")

	viper.SetEnvPrefix("PONZU")
	viper.BindPFlag("gocmd", pflags.Lookup("gocmd"))
	// bind the flags for run to environment variables, with PONZU_ prefix.
	viper.BindPFlag("port", runCmd.Flags().Lookup("port"))
	viper.BindPFlag("httpsport", runCmd.Flags().Lookup("httpsport"))
	viper.BindPFlag("devhttps", runCmd.Flags().Lookup("devhttps"))
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

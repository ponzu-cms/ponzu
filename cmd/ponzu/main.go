package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/ponzu-cms/ponzu/system/admin"
	"github.com/ponzu-cms/ponzu/system/api"
	"github.com/ponzu-cms/ponzu/system/api/analytics"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/tls"

	_ "github.com/ponzu-cms/ponzu/content"
)

var (
	usage = usageHeader + usageNew + usageGenerate +
		usageBuild + usageRun + usageUpgrade + usageVersion
	port      int
	httpsport int
	https     bool
	devhttps  bool
	cli       bool

	// for ponzu internal / core development
	dev   bool
	fork  string
	gocmd string
)

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.IntVar(&port, "port", 8080, "port for ponzu to bind its HTTP listener")
	flag.IntVar(&httpsport, "httpsport", 443, "port for ponzu to bind its HTTPS listener")
	flag.BoolVar(&https, "https", false, "enable automatic TLS/SSL certificate management")
	flag.BoolVar(&devhttps, "devhttps", false, "[dev environment] enable automatic TLS/SSL certificate management")
	flag.BoolVar(&dev, "dev", false, "modify environment for Ponzu core development")
	flag.BoolVar(&cli, "cli", false, "specify that information should be returned about the CLI, not project")
	flag.StringVar(&fork, "fork", "", "modify repo source for Ponzu core development")
	flag.StringVar(&gocmd, "gocmd", "go", "custom go command if using beta or new release of Go")
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

		case "upgrade":
			fmt.Println(usageUpgrade)
			os.Exit(0)

		case "version", "v":
			fmt.Println(usageVersion)
			os.Exit(0)
		}

	case "new":
		if len(args) < 2 {
			fmt.Println(usageNew)
			os.Exit(0)
		}

		err := newProjectInDir(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	case "generate", "gen", "g":
		if len(args) < 3 {
			fmt.Println(usageGenerate)
			os.Exit(0)
		}

		// check what we are asked to generate
		switch args[1] {
		case "content", "c":
			err := generateContentType(args[2:])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			msg := fmt.Sprintf("Generator '%s' is not implemented.", args[1])
			fmt.Println(msg)
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

		if devhttps {
			addTLS = "--devhttps"
		}

		var services string
		if len(args) > 1 {
			services = args[1]
		} else {
			services = "admin,api"
		}

		serve := exec.Command("./ponzu-server",
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--httpsport=%d", httpsport),
			addTLS,
			"serve",
			services,
		)
		serve.Stderr = os.Stderr
		serve.Stdout = os.Stdout

		err := serve.Run()
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

			for _, service := range services {
				if service == "api" {
					api.Run()
				} else if service == "admin" {
					admin.Run()
				} else {
					fmt.Println("To execute 'ponzu serve', you must specify which service to run.")
					fmt.Println("$ ponzu --help")
					os.Exit(1)
				}
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

	case "version", "v":
		// read ponzu.json value to Stdout

		p, err := ponzu(cli)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Ponzu v%s\n", p["version"])

	case "upgrade":
		// confirm since upgrade will replace Ponzu core files
		path, err := os.Getwd()
		if err != nil {
			fmt.Println("Failed to find current directory.", err)
			os.Exit(1)
		}

		fmt.Println("Only files you added to this directory, 'addons' and 'content' will be preserved.")
		fmt.Println("Upgrade this project? (y/N):")

		answer, err := getAnswer()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch answer {
		case "n", "no", "\r\n", "\n", "":
			fmt.Println("")

		case "y", "yes":
			err := upgradePonzuProjectDir(path)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		default:
			fmt.Println("Input not recognized. No upgrade made. Answer as 'y' or 'n' only.")
		}

	case "":
		fmt.Println(usage)
		fmt.Println(usageHelp)

	default:
		fmt.Println(usage)
		fmt.Println(usageHelp)
	}
}

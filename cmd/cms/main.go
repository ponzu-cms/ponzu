package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nilslice/cms/system/admin"
	"github.com/nilslice/cms/system/db"
)

var usage = `
$ cms <option> <params>

Options 

new <directory>:

	Creates a new 'cms' in the current directory, or one supplied
	as a parameter immediately following the 'new' option. Note: 'new'
	depends on the program 'git' and possibly a network connection. If there is
	no local repository to clone from at the local machine's $GOPATH, 'new' will
	attempt to clone the 'cms' package from over the network.

	Example:
	$ cms new ~/Projects/food-reviews.dev



generate, gen, g <type>:

    Generate a content type file with boilerplate code to implement
    the editor.Editable interface. Must be given one (1) parameter of
    the name of the type for the new content.

    Example:
	$ cms gen review



serve, s <service:port:tls>:

	Starts the 'cms' HTTP server for the JSON API, Admin System, or both.
	Must be given at least one (1) parameter. The segments describe 
	which services to start, in order, either 'admin' 
	(Admin System / CMS backend) or 'api' (JSON API), the port to which the 
	service is bound, and, optionally, if the server(s) should utilize TLS
	encryption (served over HTTPS), which is automatically managed using 
	Let's Encrypt (https://letsencrypt.org) 

	Example: 
	$ cms serve admin:8080:tls api:8000
	(or) 
	$ cms serve admin:8080
	(or)
	$ cms serve api:8000:tls
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
		admin.Run("8080")
		// api.Run("8000")

	case "":
		flag.PrintDefaults()
	default:
		flag.PrintDefaults()
	}
}

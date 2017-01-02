![My friend, "Gotoro"](ponzu-banner.png)

# Ponzu
Ponzu is a powerful and efficient open-source "Content-as-a-Service" system 
framework and CMS. It provides automatic, free, and secure HTTP/2 over TLS (certificates 
obtained via Let's Encrypt - https://letsencrypt.org), a useful CMS and 
scaffolding to generate content editors, and a fast HTTP API on which to build 
modern applications.

Ponzu is released under the BSD-3-Clause license (see LICENSE).
(c) 2016 Boss Sauce Creative, LLC

## Installation

```
$ go get github.com/ponzu-cms/ponzu/...
```

### Requirements
Go 1.8+

Since HTTP/2 Server Push is used, Go 1.8+ is required. However, it is not 
required of clients conntecing to a Ponzu server to make requests over HTTP/2. 

## Usage

```bash
$ ponzu [flags] command <params>
```

## Commands

### new

Creates a 'ponzu' directory, or one by the name supplied as a parameter 
immediately following the 'new' option in the $GOPATH/src directory. Note: 
'new' depends on the program 'git' and possibly a network connection. If 
there is no local repository to clone from at the local machine's $GOPATH, 
'new' will attempt to clone the 'github.com/ponzu-cms/ponzu' package from 
over the network.

Example:
```bash
$ ponzu new myProject
> New ponzu project created at $GOPATH/src/myProject
```

Errors will be reported, but successful commands retrun nothing.

---

### generate, gen, g

Generate a content type file with boilerplate code to implement
the editor.Editable interface. Must be given one (1) parameter of
the name of the type for the new content. The fields following a 
type determine the field names and types of the content struct to 
be generated. These must be in the following format:
fieldName:"T"

Example:
```bash
                   struct fields and built-in types...
                   |
                   v    
$ ponzu gen review title:"string" body:"string" rating:"int" tags:"[]string"
            ^
            |
            struct type
```

The command above will generate a file `content/review.go` with boilerplate
methods, as well as struct definition, and cooresponding field tags like:

```go
type Review struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Rating int      `json:"rating"`
	Tags   []string `json:"tags"`
}
```

The generate command will intelligently parse more sophisticated field names
such as 'field_name' and convert it to 'FieldName' and vice versa, only where 
appropriate as per common Go idioms. Errors will be reported, but successful 
generate commands retrun nothing.

---

### build

From within your Ponzu project directory, running build will copy and move 
the necessary files from your workspace into the vendored directory, and 
will build/compile the project to then be run. 

Optional flags:
- `--gocmd` sets the binary used when executing `go build` witin `ponzu` build step

Example:
```bash
$ ponzu build
(or)
$ ponzu --gocmd=go1.8beta2 build # useful for testing
```

Errors will be reported, but successful build commands return nothing.

---

### run

Starts the HTTP server for the JSON API, Admin System, or both.
The segments, separated by a comma, describe which services to start, either 
'admin' (Admin System / CMS backend) or 'api' (JSON API), and, optionally, 
if the server should utilize TLS encryption - served over HTTPS, which is
automatically managed using Let's Encrypt (https://letsencrypt.org) 

Optional flags:
- `--port` sets the port on which the server listens for requests [defaults to 8080]
- `--https` enables auto HTTPS management via Let's Encrypt (port is always 443)
- `--devhttps` generates self-signed SSL certificates for development-only (port is 10443)

Example: 
```bash
$ ponzu run
(or)
$ ponzu --port=8080 --https run admin,api
(or) 
$ ponzu run admin
(or)
$ ponzu --port=8888 run api
(or)
$ ponzu --devhttps run
```
Defaults to `$ ponzu --port=8080 run admin,api` (running Admin & API on port 8080, without TLS)

*Note:* 
Admin and API cannot run on separate processes unless you use a copy of the
database, since the first process to open it receives a lock. If you intend
to run the Admin and API on separate processes, you must call them with the
'ponzu' command independently.

---


## Contributing

1. Checkout branch ponzu-dev
2. Make code changes
3. Test changes to ponzu-dev branch
    - make a commit to ponzu-dev
    - to manually test, you will need to use a new copy (ponzu new path/to/code), but pass the --dev flag so that ponzu generates a new copy from the ponzu-dev branch, not master by default (i.e. `$ponzu --dev new /path/to/code`)
    - build and run with $ ponzu build and $ ponzu run
4. To add back to master: 
    - first push to origin ponzu-dev
    - create a pull request 
    - will then be merged into master

_A typical contribution workflow might look like:_
```bash
# clone the repository and checkout ponzu-dev
$ git clone https://github.com/ponzu-cms/ponzu path/to/local/ponzu # (or your fork)
$ git checkout ponzu-dev

# install ponzu with go get or from your own local path
$ go get github.com/ponzu-cms/ponzu/...
# or
$ cd /path/to/local/ponzu 
$ go install ./...

# edit files, add features, etc
$ git add -A
$ git commit -m 'edited files, added features, etc'

# now you need to test the feature.. make a new ponzu project, but pass --dev flag
$ ponzu --dev new /path/to/new/project # will create $GOPATH/src/path/to/new/project

# build & run ponzu from the new project directory
$ cd /path/to/new/project
$ ponzu build && ponzu run

# push to your origin:ponzu-dev branch and create a PR at ponzu-cms/ponzu
$ git push origin ponzu-dev
# ... go to https://github.com/ponzu-cms/ponzu and create a PR
```

**Note:** if you intend to work on your own fork and contribute from it, you will
need to also pass `--fork=path/to/your/fork` (using OS-standard filepath structure),
where `path/to/your/fork` _must_ be within `$GOPATH/src`, and you are working from a branch
called `ponzu-dev`. 

For example: 
```bash
# ($GOPATH/src is implied in the fork path, do not add it yourself)
$ ponzu --dev --fork=github.com/nilslice/ponzu new /path/to/new/project
```


## Credits
- [golang.org/x/text/unicode/norm](https://golang.org/x/text/unicode/norm)
- [golang.org/x/text/transform](https://golang.org/x/text/transform)
- [golang.org/x/crypto/bcrypt](https://golang.org/x/crypto/bcrypt)
- [github.com/nilslice/jwt](https://github.com/nilslice/jwt)
- [github.com/nilslice/rand](https://github.com/nilslice/rand)
- [github.com/nilslice/email](https://github.com/nilslice/email)
- [github.com/gorilla/schema](https://github.com/gorilla/schema)
- [github.com/satori/go.uuid](https://github.com/satori/go.uuid)
- [github.com/tidwall/gjson](https://github.com/tidwall/gjson)
- [github.com/boltdb/bolt](https://github.com/boltdb/bolt)
- [github.com/sluu99/um](https://github.com/sluu99/um)
- [Materialnote Editor](http://www.web-forge.info/projects/materialNote)
- [Materialize.css](http://materialize.css)
- [jQuery](https://jquery.com/)
- [Chart.js](http://www.chartjs.org/)


### Logo
The Go gopher was designed by Renee French. (http://reneefrench.blogspot.com)
The design is licensed under the Creative Commons 3.0 Attributions license.
Read this article for more details: http://blog.golang.org/gopher

The Go gopher vector illustraition by Hugo Arganda [@argandas](https://twitter.com/argandas) (http://about.me/argandas)

"Gotoro", the sushi chef, is a modification of Hugo Arganda's illustration by Steve Manuel (https://github.com/nilslice).

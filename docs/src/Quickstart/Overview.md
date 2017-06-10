### Quickstart Steps
1) Install [Go 1.8+](https://golang.org/dl/)

2) Install Ponzu CLI:
```bash
$ go get github.com/ponzu-cms/ponzu/…
```

3) Create a new project (path is created in your GOPATH):
```bash
$ ponzu new github.com/nilslice/reviews
```

4) Enter your new project directory:
```bash
$ cd $GOPATH/src/github.com/nilslice/reviews
```

5) Generate content type file and boilerplate code (creates `content/review.go`):
```bash
$ ponzu generate content review title:"string" author:"string" rating:"float64" body:"string":richtext website_url:"string" items:"[]string" photo:string:file`
```

6) Build your project:
```bash
$ ponzu build
```

7) Run your project with defaults:
```bash
$ ponzu run
```

8) Open browser to [`http://localhost:8080/admin`](http://localhost:8080/admin)

### Notes
- One-time initialization to set configuration
- All fields can be changed in Configuration afterward

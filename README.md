# Ponzu
## (WIP)

Creating a "API-first" CMS thing for Go developers. Not quite ready for use, but would be happy 
to get thoughts/feedback!

## Installation

```
$ go get github.com/bosssauce/ponzu/...
```

## Contributing

1. Checkout branch ponzu-dev
2. Make code changes
3. Commit change to ponzu-dev branch (I know, a little unnatural. Advice gladly accepted.)
    - to manually test, you will need to use a new copy (ponzu new path/to/code), but pass the --dev flag so that ponzu generates a new copy from the ponzu-dev branch, not master by default (i.e. $ponzu --dev new /path/to/code)
    - build and run with $ ponzu build and $ ponzu run
4. To add back to master: 
    - first push to origin ponzu-dev
    - create a pull request 
    - will then be merged into master

_A typical contribution workflow might look like:_
```bash
# clone the repository and checkout ponzu-dev
$ git clone https://github.com/bosssauce/ponzu 'path/to/local/ponzu'
$ git checkout ponzu-dev

# install ponzu with go get or from your own local path
$ go get github.com/bosssauce/ponzu/...
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

# push to your origin:ponzu-dev branch and create a PR at bosssauce/ponzu
$ git push origin ponzu-dev
# ... go to https://github.com/bosssauce/ponzu and create a PR
```

**Note:** if you intend to work on your own fork and contribute from it, you will
need to also pass `--fork=path/to/your/fork` (using OS-standard filepath structure),
where `path/to/your/fork` _must_ be within `$GOPATH/src`. 

For example: 
```bash
# ($GOPATH/src is implied in the fork path, do not add it yourself)
$ ponzu --dev --fork=github.com/nilslice/ponzu new /path/to/new/project
```



## Credits
- [golang.org/x/text/transform](https://golang.org/x/text/transform)
- [golang.org/x/text/unicode/norm](https://golang.org/x/text/unicode/norm)
- [github.com/nilslice/jwt](https://github.com/nilslice/jwt)
- [github.com/nilslice/rand](https://github.com/nilslice/rand)
- [golang.org/x/crypto/bcrypt](https://golang.org/x/crypto/bcrypt)
- [github.com/boltdb/bolt](https://github.com/boltdb/bolt)
- [github.com/gorilla/schema](https://github.com/gorilla/schema)
- [github.com/sluu99/um](https://github.com/sluu99/um)
- [Materialize.css](http://materialize.css)
- [Materialnote Editor](http://www.web-forge.info/projects/materialNote)
- [jQuery](https://jquery.com/)

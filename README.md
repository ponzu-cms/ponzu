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

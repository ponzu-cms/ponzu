# TODO: remove from here and add to wki

## Ponzu Docker build

Ponzu is distributed as a docker image **NEEDS LINK**, which aids in docker deployment. The Dockerfile in this directory is used by ponzu to generate the docker image which contains the ponzu executable.

### The following are convient commands during development of Ponzu.

#### Build the docker image. Run from the root of the project.
```bash
# from the root of ponzu:
docker build -t ponzu-dev
```

#### Start the image, share the local directory and pseudo terminal (tty) into for debugging:
```bash
docker run -v $(pwd):/go/src/github.com/ponzu-cms/ponzu -it ponzu-dev
pwd # will output the go src directory for ponzu
ponzu version # will output the ponzu version
# make an edit on your local and rebuild
go get github.com/ponzu-cms/ponzu/...
```
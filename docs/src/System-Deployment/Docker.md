## Ponzu Docker build

Ponzu is distributed as a [docker image](https://hub.docker.com/r/ponzu/ponzu/), 
which aids in ponzu deployment. The Dockerfile in this directory is used by Ponzu 
to generate the docker image which contains the ponzu executable.

If you are deploying your own Ponzu project, you can write a new Dockerfile that
is based from the `ponzu/ponzu` image of your choice. For example:
```docker
FROM ponzu/ponzu:latest

# your project set up ...
# ...
# ...
```

### The following are convenient commands during development of Ponzu core:

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
go install ./...
```

Special thanks to [**@krismeister**](https://github.com/krismeister) for contributing this!
## Ponzu Docker build

Ponzu is distributed as a docker image **NEEDS LINK**, which aids in docker deployment. The Dockerfile in this directory is used by ponzu to generate the docker image which contains the ponzu executable.

### The following are convient commands during development of Ponzu.

#### Build the docker image. Run from the root of the project.
```bash
# from the root of ponzu:
docker build -f docker/Dockerfile -t ponzu-dev .
```

#### Start the image and SSH into for debugging:
```bash
docker run -it ponzu-dev
pwd #will output /ponzu
ponzu version #will output the ponzu version
```
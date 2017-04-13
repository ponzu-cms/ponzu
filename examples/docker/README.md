## Example of Running Ponzu with Docker

This docker-compose contains 2 docker containers:

1. web - a basic nginx front end, with Javascript ajax accessing the /api
2. admin - an example ponzu container

The admin contain is based on the official Ponzu Docker image **NEEDS LINK**

### Running the example

```bash
# build the containers
docker-compose build

#start the containers in the background
docker-compose start -d
```
Visit the http://localhost:3000/admin/init to configure Ponzu.

Stop the on containers:
```
docker-compose stop
```

### Web Container
This nginx web container takes any incoming requests and if it matches `/api*` or `/admin*` it will then route it to the exposed ports :8080 on the ponzu container

### Ponzu - Admin Container

The ponzu container has a small [startup script](./admin/start_admin_.sh) which symlinks and logs or database files into a Docker Volume. If you need access to the ponzu logs, it is exposed a Volume.

#### Accessing the Ponzu terminal in Development

Make sure you have no running docker images `docker ps`. Then use the following command to start an interactive shell into the docker container.

```bash
# make sure the containers are built
docker-compose build

# start and tty into the container
docker run -v $(pwd)/admin:/go/src/project -it docker_admin

# run a ponzu command
ponzu generate content message title:"string" description:"string"

```

After the above your new `message.go` model is now available in your local filesystem. Use `docker-compose up -d` to see the new model in the admin.

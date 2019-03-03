Heroku describes itself as "a platform as a service based on a managed container system, with integrated data services and a powerful ecosystem, for deploying and running modern apps." In the end, it's a great place to play with tech on a powerful platform that offers a generous free plan.

## First, the heroku.yml

First things first, you'll need to have the [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed.

After you have your Ponzu application working the way you'd like. From the root of your directory create a `heroku.yml` file. Here we are going to tell Heroku that we want to use a container and utilize the `Dockerfile` to do the real work. The `heroku.yml` file should look like this:

```yaml
build:
  docker:
    web: Dockerfile
```

For more information about the `heroku.yml` file see [Building Docker Images with heroku.yml](https://devcenter.heroku.com/articles/build-docker-images-heroku-yml).

## Second, the Dockerfile

Next up is the `Dockerfile` also in the root of your Ponzu project. I'll point out what needs to be changed in the `Dockerfile` but it should resemble this:

```docker
FROM golang:1.9

ENV PONZU_GITHUB github.com/ponzu-cms/ponzu
# Change this to your repo HTTPS link
ENV PROJECT_GITHUB https://github.com/itzsaga/bookshelf.git
# Change this to your project root
ENV PROJECT_ROOT $GOPATH/src/github.com/itzsaga/bookshelf

RUN go get $PONZU_GITHUB/...

RUN mkdir -p $PROJECT_ROOT

WORKDIR $PROJECT_ROOT
RUN git clone $PROJECT_GITHUB .

RUN ponzu build
CMD ponzu run --port=${PORT} --bind=0.0.0.0
```

The two lines that need to be changed are in regards to your project code. The Heroku Docker container is going to need to get the project code from somewhere to run it. I've found the simplest way is via a Github repository.

The grand idea is, setup a container with Go on it, install Ponzu, clone the project code into the proper directory, go into that directory, build the project, and run the server.

After the `Dockerfile` is setup you are ready to deploy.

## Third, deploying to Heroku

If you're familiar with the Heroku CLI then you might be thinking that after you've created the `heroku.yml` and `Dockerfile` you can just follow the usual steps. However, that is not the case. There is one extra step in the process. To get this running the following steps will need to be followed:

```bash
git add heroku.yml
git add Dockerfile
git commit -m "Add heroku files for deployment"
heroku stack:set container
git push heroku master
```

After that you should see a bunch of things going on in your terminal. That's good! However, you won't know if everything worked correctly strictly from what you see in the terminal. This is because the terminal is only reporting on if the container builds correctly. Not if the server that's started inside that container is working correctly. So go on over to the Heroku dashboard, click on the app that's running Ponzu, click the "More" button and "View Logs" to make sure everything is working as it should be.

Congrats! You're now up and running with Ponzu on Heroku.

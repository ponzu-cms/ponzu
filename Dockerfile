
# Base our image on an official, minimal image of our preferred golang
FROM golang:1.8.1

# Note: The default golang docker image, already has the GOPATH env variable set.
# GOPATH is located at /go
ENV GO_SRC $GOPATH/src
ENV PONZU_GITHUB github.com/ponzu-cms/ponzu
ENV PONZU_ROOT $GO_SRC/$PONZU_GITHUB

# Consider updating package in the future. For instance ca-certificates etc.
# RUN apt-get update -qq && apt-get install -y build-essential

# Make the ponzu root directory
RUN mkdir -p $PONZU_ROOT

# All commands will be run inside of ponzu root
WORKDIR $PONZU_ROOT

# Copy the ponzu source into ponzu root.
COPY . .

# the following runs the code inside of the $GO_SRC/$PONZU_GITHUB directory
RUN go get -u $PONZU_GITHUB...

# Define the scripts we want run once the container boots
# CMD [ "" ]

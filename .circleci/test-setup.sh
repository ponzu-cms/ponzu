#!/bin/bash

# Set up test environment

set -ex 

# Install Ponzu CMS
go get github.com/ponzu-cms/ponzu/...

# test install
ponzu

PROJ_PATH=github.com/ponzu-cms/ci/test-project
# create a project and generate code
if [ $CIRCLE_BRANCH = "ponzu-dev" ]; then
        # ensure we have the latest from ponzu-dev branch
        cd /go/src/github.com/ponzu-cms/ponzu
        git checkout ponzu-dev
        git pull origin ponzu-dev

        # install ponzu-dev branch of CLI
        go install ./...

        # create new project using the ponzu-dev branch
        ponzu new --dev ${PROJ_PATH}
else 
        ponzu new ${PROJ_PATH}
fi

cd /go/src/${PROJ_PATH}

ponzu gen content person name:string hashed_secret:string
ponzu gen content message from:@person,hashed_secret to:@person,hashed_secret

# build and run dev http/2 server with TLS
ponzu build


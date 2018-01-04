#!/bin/bash

# Test that the configuration runs.

set -ex

cd /go/src/github.com/ponzu-cms/ci/test-project

ponzu run --dev-https &        

# Smoke test
sleep 2
curl -Lk https://localhost:10443/admin | grep https://www.bosssauce.it

# Run unit tests
touch cookies

# Create initial admin user
curl -v --cookie-jar cookies \
    -d "name=Test&domain=localhost&email=test@ponzu-cms.org&password=ponzu" \
    -X POST localhost:8080/admin/init

#Test that content types were generated
curl -b cookies -c cookies http://localhost:8080/admin/contents?type=Person \
    | grep Person

curl -b cookies -c cookies http://localhost:8080/admin/contents?type=Message \
    | grep Message


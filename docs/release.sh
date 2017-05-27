#! /bin/bash
set -e

echo "---- [release] Building Docs ----"
docker run --rm -it -p 8000:8000 -v `pwd`:/docs squidfunk/mkdocs-material build

cp CNAME ./build

git add -A
git commit -m "$1" 

echo "---- [release] Push: Master ----"
git push origin master

echo "---- [release] Push: Build ----"
git subtree push --prefix build origin gh-pages

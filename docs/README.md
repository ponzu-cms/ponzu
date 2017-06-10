# Ponzu CMS + Server Framerwork Docs

## Contributing

Documentation contributions are welcome and appreciated. If you find something 
lacking in documentation or have submitted a PR that is being merged into master, 
please help everyone out and write some docs! 

**Note:** Docker is required to follow these instructions, but you can also use
MkDocs natively, [see details here](http://www.mkdocs.org/#installation). Ponzu
docs use the "Material" [theme](http://squidfunk.github.io/mkdocs-material/).


Here is how to run a local docs server and build them for release:

1. Clone this repository
```bash
$ git clone https://github.com/ponzu-cms/docs.git
``` 
2. Start the development server which will watch and auto-build the docs
```bash
$ docker run --rm -it -p 8000:8000 -v `pwd`:/docs squidfunk/mkdocs-material
``` 
3. Submit a PR with your changes. If you run the build step, please do not add it to the PR.

**Thank you!**

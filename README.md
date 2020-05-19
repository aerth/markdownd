![markdownd](https://github.com/aerth/markdownd/blob/master/docs/markdownd.png?raw=true)

`markdownd [flags] <directory>`

`markdownd -toc -header theme/header.html -footer theme/footer.html .`

`markdownd -index=gen .`

[![Go Report Card](https://goreportcard.com/badge/github.com/aerth/markdownd)](https://goreportcard.com/report/github.com/aerth/markdownd) 
[![Build Status](https://travis-ci.org/aerth/markdownd.svg?branch=master)](https://travis-ci.org/aerth/markdownd) 

## Markdown Server

  * tries markdown file (.md) in .html request (/index.html tries /index.md first)
  * will serve .html if exists
  * serves static files and downloads if not .html or .md
  * optional indexing (default: off, use -index=gen or -index=README.md)
  * no symlinks
  * no `../` paths
  * raw markdown source requests ( example: `GET /index.md?raw` )
  * custom index page (use flag: `-index README.md`)
  * generates table of contents with `-toc` flag
  * themed html with `-header` and `-footer` flag
  * now with syntax highlighting (use flag: `-syntax`)

## Usage

  * `GET /` will show a 404 unless -index flag is used (-index=gen to generate)
  * `GET /README.md` or `GET /README.html` will process the markdown file and serve HTML.
  * `GET /README.md?raw` will serve raw markdown source
  * To generate index page (with links to files), use `-index=gen`
  * To serve custom `index.md`, use `-index=index.md`

#### Example use case: live preview your git repository's README.md

From your project repository that contains a README.md file, run markdownd like so:

```
markdownd -index=README.md .
```

And visit http://localhost:8080/ in your browser

## Installation

### Compile using Go (from any directory)

```
git clone https://github.com/aerth/markdownd
cd markdownd
make && sudo make install
```

If you don't want to install the server system-wide, or you don't have root privileges, replace last line with:

```
make && make install INSTALLDIR=$HOME/bin
```

### Or using legacy go get

```
GOFLAGS=-tags=netgo,osusergo GOBIN=$HOME/bin go get -v github.com/aerth/markdownd
```

### Download binary for your OS (old versions)

[Latest Release](https://github.com/aerth/markdownd/releases/latest)

Consider installing [go](https://golang.org/dl) and building from source,
Its fast and easy.

## Docker

When using the docker image, markdownd servest the /opt directory,
and exposes port 8080.

You will want to share a directory into /opt and forward the port
(using docker's `-v` and `-p` flags).

For example (modify `$PWD/docs` and `8888` to suit your needs):

`docker run -it -v $PWD/docs:/opt -p 8888:8080 aerth/markdownd`

## Free and Open Source

	The MIT License (MIT)
	
	Copyright (c) 2017-2020  aerth <aerth@riseup.net>
	
	Permission is hereby granted, free of charge, to any person obtaining a 
	copy of this software and associated documentation files (the 
	"Software"), to deal in the Software without restriction, including 
	without limitation the rights to use, copy, modify, merge, publish, 
	distribute, sublicense, and/or sell copies of the Software, and to 
	permit persons to whom the Software is furnished to do so, subject to 
	the following conditions:
	
	The above copyright notice and this permission notice shall be included 
	in all copies or substantial portions of the Software.
	
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS 
	OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF 
	MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. 
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY 
	CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, 
	TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE 
	SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
	
## Contributing:

  * pull requests welcome
  * bugs/issues/features very welcome
  * please 'gofmt -w -l -s' before commits

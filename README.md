![markdownd](https://github.com/aerth/markdownd/blob/master/docs/markdownd.png?raw=true)

`./md [flags] <directory>`

`./markdownd -toc -header theme/header.html -footer theme/footer.html .`

[![Go Report Card](https://goreportcard.com/badge/github.com/aerth/markdownd)](https://goreportcard.com/report/github.com/aerth/markdownd) 
[![Build Status](https://travis-ci.org/aerth/markdownd.svg?branch=master)](https://travis-ci.org/aerth/markdownd) 

## serves:

  * tries markdown file (.md) in .html request (/index.html tries /index.md first)
  * will serve .html if exists
  * serves static files and downloads if not .html or .md
  * no indexing
  * no symlinks
  * no `../` paths
  * raw markdown requests ( example: `GET /index.md?raw` )
  * custom index page (use flag: `-index README.md`)
  * generates table of contents with `-toc` flag
  * themed html with `-header` and `-footer` flag

## docker

example launch code, modify `$PWD/docs` and `8888` to suit your needs

  * `docker run -it -v $PWD/docs:/opt -p 8888:8080 aerth/markdownd`

## free:

	The MIT License (MIT)
	
	Copyright (c) 2017  aerth <aerth@riseup.net>
	
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
	
## contributing:

  * pull requests welcome
  * bugs/issues/features very welcome
  * please 'gofmt -w -l -s' before commits

## logging:

	2017/05/08 14:14:05 5577006791947779410 [::1]:58338 GET / -> docs/index.md
	2017/05/08 14:14:05 5577006791947779410 serving markdown: /home/aerth/go/src/github.com/aerth/markdownd/docs/index.md
	2017/05/08 14:14:05 5577006791947779410 closed
	2017/05/08 14:14:05 8674665223082153551 [::1]:58338 GET /markdownd.png -> docs/markdownd.png
	2017/05/08 14:14:05 8674665223082153551 serving image/png: /home/aerth/go/src/github.com/aerth/markdownd/docs/markdownd.png
	2017/05/08 14:14:05 8674665223082153551 closed
	2017/05/08 14:14:07 6129484611666145821 [::1]:58338 GET /index.html -> docs/index.html
	2017/05/08 14:14:07 6129484611666145821 docs/index.html -> docs/index.md
	2017/05/08 14:14:07 6129484611666145821 serving markdown: /home/aerth/go/src/github.com/aerth/markdownd/docs/index.md
	2017/05/08 14:14:07 6129484611666145821 closed
	2017/05/08 14:14:08 4037200794235010051 [::1]:58338 GET /test.html -> docs/test.html
	2017/05/08 14:14:08 4037200794235010051 serving raw html: /home/aerth/go/src/github.com/aerth/markdownd/docs/test.html
	2017/05/08 14:14:08 4037200794235010051 closed
	2017/05/08 14:14:09 3916589616287113937 [::1]:58338 GET / -> docs/index.md
	2017/05/08 14:14:09 3916589616287113937 serving markdown: /home/aerth/go/src/github.com/aerth/markdownd/docs/index.md
	2017/05/08 14:14:09 3916589616287113937 closed

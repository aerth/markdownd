FROM golang
MAINTAINER aerth <aerth@riseup.net>
RUN env CGO_ENABLED=0 go get -v -x -ldflags='-w -s' github.com/aerth/markdownd
CMD markdownd /opt

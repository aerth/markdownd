# Copyright 2017-2020 aerth <aerth@riseup.net>
# Markdownd (latest source: https://github.com/aerth/markdownd)

# build static linked markdownd
export GOFLAGS=-tags=netgo,osusergo

markdownd: *.go
	go build -o $@ -v
clean:
	rm -f ./markdownd
install: markdownd
	install markdownd /usr/local/bin/
test:
	go test -v ./...
docker-build:
	docker build -t aerth/markdownd --network host .
docker-run: 
	docker run -it -v "${PWD}:/opt" -p "127.0.0.1:8080:8080" aerth/markdownd

.PHONY += clean
.PHONY += install
.PHONY += test
.PHONY += docker-build
.PHONY += docker-run


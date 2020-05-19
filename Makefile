# Copyright 2017-2020 aerth <aerth@riseup.net>
# Markdownd (latest source: https://github.com/aerth/markdownd)

INSTALLDIR ?= /usr/local/bin/

# build static linked markdownd
export GOFLAGS=-tags=netgo,osusergo

markdownd: *.go
	go build -o $@ -v
clean:
	rm -f ./markdownd
install: markdownd
	@install markdownd ${INSTALLDIR} && echo "installation complete"
test:
	go test -v ./...
docker-build:
	docker build -t aerth/markdownd --network host .
docker-run: 
	@echo Serving http://localhost:8888
	docker run -it \
	  -v "${PWD}:/opt" \
	  -p "127.0.0.1:8888:8080" \
	  aerth/markdownd -http=:8080 -index=README.md /opt

.PHONY += clean
.PHONY += install
.PHONY += test
.PHONY += docker-build
.PHONY += docker-run


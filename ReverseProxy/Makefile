.PHONY: build get

GOC=go
INSTALL=/usr/bin/install -D

VENDORPATH:=${PWD}/_vendor
GOPATH:=$(VENDORPATH):${PWD}

VERSION=1.0
REVISION=$(shell git rev-parse --short HEAD)

all: build

install: build
	$(INSTALL) --mode 0755 bin/docserver $(DESTDIR)/usr/bin/docserver

clean:
	rm -rf ./bin

build:
	mkdir -p bin
	GOPATH=$(GOPATH) $(GOC) build -ldflags="-X main.version=${VERSION}-${REVISION}" -o bin/reverseproxy proxy.go

get:
	mkdir -p ${VENDORPATH}
	GOPATH=$(VENDORPATH) $(GOC) get -d -v  github.com/BurntSushi/toml
	find _vendor/src/ -depth \( -name ".git" -or -name ".svn" -or -name ".hg" -or -name ".bzr" \) -exec rm -rf {} \;	






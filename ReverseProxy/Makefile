VERSION=1.0
REVISION=$(shell git rev-parse --short HEAD)

build:
	go build -ldflags="-X main.version ${VERSION}-${REVISION}"

get:
	go get github.com/BurntSushi/toml

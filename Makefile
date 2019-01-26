BINARY_NAME=trek
URL=https://releases.hashicorp.com/nomad/0.8.6/nomad_0.8.6_darwin_amd64.zip

all: build_tools dependencies build release

build_tools:
	go get github.com/goreleaser/goreleaser

dependencies:
	go get github.com/jroimartin/gocui
	go get github.com/nsf/termbox-go
build:
	go build

release:
	goreleaser --rm-dist

clean:
	rm $(BINARY_NAME)

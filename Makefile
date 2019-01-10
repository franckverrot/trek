BINARY_NAME=trek
URL=https://releases.hashicorp.com/nomad/0.8.6/nomad_0.8.6_darwin_amd64.zip

trek:
	go build

.PHONY: ui
ui: trek
	./$(BINARY_NAME)

no-ui: trek
	./$(BINARY_NAME) -ui=false example

clean:
	rm $(BINARY_NAME)

.PHONY: prepare
prepare:
	rm -f nomad
	curl $(URL) -o nomad.zip
	unzip nomad.zip
	rm nomad.zip
	./nomad agent -dev

.PHONY: demo
demo:
	./nomad job run tests/example34.nomad
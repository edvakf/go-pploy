hash := $(shell git rev-parse --verify HEAD)

.PHONY: build
build: go-pploy

go-pploy: main.go $(wildcard web/*.go) web/assets.go
	go get
	go build -ldflags "-X main.GitCommit=${hash}"

.PHONY: cross-build
cross-build: main.go $(wildcard web/*.go) web/assets.go
	go get github.com/mitchellh/gox
	$(GOPATH)/bin/gox -os="darwin linux" -arch="amd64" -ldflags="-X main.GitCommit=${hash}" -output "pkg/{{.Dir}}_{{.OS}}_{{.Arch}}"

web/assets.go: assets/bootstrap assets/izitoast assets/bundle.js $(wildcard assets/*)
	go get github.com/jessevdk/go-assets-builder
	$(GOPATH)/bin/go-assets-builder -p web assets/ > $@

assets/bundle.js: $(wildcard svelte/*.svelte) svelte/main.js
	npm run build

assets/bootstrap: node_modules
	rsync -a node_modules/bootstrap/dist/ assets/bootstrap/
	touch assets/bootstrap

assets/izitoast: node_modules
	rsync -a node_modules/izitoast/dist/ assets/izitoast/
	touch assets/izitoast

node_modules:
	npm install
	touch node_modules

.PHONY: clean
clean:
	rm web/assets.go
	rm assets/bundle.*
	rm -r assets/bootstrap/*
	rm -r assets/izitoast/*
	rm -r node_modules
	rm go-pploy

.PHONY: test
test:
	go test -v ./unbuffered

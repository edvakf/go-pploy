hash := $(shell git rev-parse --verify HEAD)

.PHONY: test components bootstrap izitoast go-assets-builder svelte clean dep gox

# `make` builds pploy
pploy: vendor main.go $(wildcard web/*.go) web/assets.go
	go build -ldflags "-X main.hash=${hash}"

gox: vendor main.go $(wildcard web/*.go) web/assets.go
	gox -os="linux darwin" -arch="amd64" -ldflags="-X main.hash=${hash}"

# please run `make prepare` before first build
prepare: go-assets-builder dep vendor node_modules svelte

test: web/assets.go
	go test -v ./...

web/assets.go: $(wildcard assets/*) components bootstrap izitoast
	go-assets-builder -p web assets/ > $@

components: $(wildcard svelte/*.html)
	svelte compile -m -i svelte -o assets/components

bootstrap: node_modules
	rsync -a node_modules/bootstrap/dist/ assets/bootstrap/

izitoast: node_modules
	rsync -a node_modules/izitoast/dist/ assets/izitoast/

node_modules:
	npm install

vendor:
	dep ensure

go-assets-builder:
	go get -u github.com/jessevdk/go-assets-builder

dep:
	go get -u github.com/golang/dep/cmd/dep

# install `svelte` command
svelte:
	npm install -g svelte-cli

clean:
	rm web/assets.go
	rm assets/components/*
	rm -r assets/bootstrap/*
	rm -r node_modules
	rm -r vendor
	rm pploy

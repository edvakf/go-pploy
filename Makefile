hash := $(shell git rev-parse --verify HEAD)

.PHONY: test components bootstrap izitoast go-assets-builder svelte clean gox

# `make` builds go-pploy
go-pploy: main.go $(wildcard web/*.go) web/assets.go
	go build -ldflags "-X main.GitCommit=${hash}"

gox: main.go $(wildcard web/*.go) web/assets.go
	gox -os="darwin linux" -arch="amd64" -ldflags="-X main.GitCommit=${hash}"

# please run `make prepare` before first build
prepare: go-assets-builder node_modules

test: web/assets.go
	go test -v ./...

web/assets.go: $(wildcard assets/*) components bootstrap izitoast
	go-assets-builder -p web assets/ > $@

components: $(wildcard svelte/*.html)
	npx svelte compile svelte -f es -o assets/components

bootstrap: node_modules
	rsync -a node_modules/bootstrap/dist/ assets/bootstrap/

izitoast: node_modules
	rsync -a node_modules/izitoast/dist/ assets/izitoast/

node_modules:
	npm install

go-assets-builder:
	go get -u github.com/jessevdk/go-assets-builder

clean:
	rm web/assets.go
	rm assets/components/*
	rm -r assets/bootstrap/*
	rm -r node_modules
	rm pploy

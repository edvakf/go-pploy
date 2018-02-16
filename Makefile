# `make` builds pploy
pploy: vendor main.go $(wildcard web/*.go) web/assets.go
	go build -o pploy .

# please run `make prepare` before first build
prepare: node_modules vendor go-assets-builder

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

clean:
	rm web/assets.go
	rm assets/components/*
	rm -r assets/bootstrap/*
	rm -r node_modules
	rm -r vendor
	rm pploy

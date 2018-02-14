pploy: main.go assets.go
	go build -o pploy .

assets.go: assets/index.html assets/index.js components bootstrap
	go-assets-builder assets/ > $@

components: $(wildcard svelte/*.html)
	svelte compile -m -i svelte -o assets/components

bootstrap: node_modules
	rsync -a node_modules/bootstrap/dist/ assets/bootstrap/

node_modules:
	npm install

vendor:
	dep ensure

install: node_modules vendor
	go get -u github.com/jessevdk/go-assets-builder

clean:
	rm assets.go
	rm assets/components/*
	rm -r assets/bootstrap/*
	rm -r node_modules
	rm -r vendor
	rm pploy

pploy: main.go assets.go
	go build -o pploy .

assets.go: assets/index.html assets/index.js components
	go-assets-builder assets/ > $@

components: $(wildcard svelte/*.html)
	svelte compile -m -i svelte -o assets/components

install:
	dep ensure
	npm install -g svelte-cli
	go get -u github.com/jessevdk/go-assets-builder

clean:
	rm assets.go
	rm assets/components/*

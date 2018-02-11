pploy: main.go assets.go
	go build -o pploy .

assets.go: assets/index.html assets/index.js components
	go-assets-builder assets/ > $@

components: $(wildcard svelte/*.html)
	# svelte compile $@ > assets/components/$(notdir $(basename $@))
	svelte compile -i svelte -o assets/components

# assets/components/App.js:
# 	svelte compile svelte/App.html > $@
# 	#svelte compile --format iife svelte/App.html > $@
#
# assets/components/Sidebar.js:
# 	svelte compile svelte/Sidebar.html > $@
#
# assets/components/Projects.js:
# 	svelte compile svelte/Projects.html > $@
#
# assets/components/Lock.js:
# 	svelte compile svelte/Lock.html > $@

install:
	dep ensure
	npm install -g svelte-cli
	go get -u github.com/jessevdk/go-assets-builder

clean:
	rm assets.go
	rm assets/components/*

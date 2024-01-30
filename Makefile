generate:
	@npx tailwindcss -i server/main.css -o server/static/tailwind.css
	@templ generate ./...
	@weaver generate ./...

build: generate
	go build -o .bin/templweaver main.go

run: build
	@.bin/templweaver


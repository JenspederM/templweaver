generate:
	@echo "Generating code..."
	@npx tailwindcss -i ./frontend/main.css -o ./frontend/static/tailwind.css
	@templ generate ./...
	@weaver generate ./...

build: generate
	@echo "Building..."
	@go build -o  ./main.go
	@echo "Built to ./bin/templweaver"
	
run: build
	@echo "Running..."
	@weaver single deploy weaver.toml

test: build
	@echo "Testing..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf ./bin

.PHONY: generate build run test clean
.DEFAULT_GOAL := build
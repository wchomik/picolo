BINARY_NAME = picolo
MODULE = github.com/wchomik/picolo
LDFLAGS = -s -w

.PHONY: build build-linux build-darwin build-windows clean install

build: ## Build for current platform
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) .

build-linux: ## Build static Linux binary
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 .

build-darwin: ## Build macOS binary
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 .

build-windows: ## Build Windows binary
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-windows-amd64.exe .

clean: ## Remove built binary
	rm -f $(BINARY_NAME)* *.exe

install: build ## Install to /usr/local/bin
	sudo mv $(BINARY_NAME) /usr/local/bin/

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

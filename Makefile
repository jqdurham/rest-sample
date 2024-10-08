help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


pre-commit: lint govuln test ## run linter, vulnerability scanner and unit tests

test: ## run unit tests, check for race conditions and report code coverage
	go test --race -coverprofile cover.out ./...
	rm cover.out

govuln: ## run govulncheck to scan for known vulnerabilities in Go or imported packages
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

lint: ## run linters using golangci-lint configuration file (default lookup)
	golangci-lint cache clean
	golangci-lint run -v ./...

run: ## run the RESTful server
	CGO_ENABLED=false go run ./cmd/rest/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o rest-sample cmd/rest/main.go
	scp ./rest-sample rest-sample:~/rest-sample
	rm ./rest-sample


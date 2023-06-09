PROJECT_NAME := "cwgoctl"
PKG := "github.com/CoverWhale/coverwhale-go"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
VERSION := $$(git describe HEAD)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

cwgoctl: ## Builds the binary on the current platform
	go build -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'" -o $(PROJECT_NAME)

docs: cwgoctl ## Builds the cli documentation
	mkdir -p docs
	./$(PROJECT_NAME) docs

clean: ## Remove previous build
	git clean -fd
	git clean -fx
	git reset --hard

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

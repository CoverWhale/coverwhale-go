// Copyright 2025 Sencillo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tpl

func Makefile() []byte {
	return []byte(`PROJECT_NAME := "{{ .Name }}"
PKG := "{{ .Module }}"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
LOCAL_VERSION := $(shell if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then git describe --exact-match --tags HEAD 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)"; else echo "dev"; fi)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPRIVATE=github.com/SencilloDev

# Version and image repo are overriden by the ci pipeline
VERSION=x.x.x
IMAGE_REPO=local/${shell basename ${PWD}}
IMAGE=${IMAGE_REPO}:${VERSION}
TEST_IMAGE:=${IMAGE_REPO}-test:${VERSION}

.PHONY: all build docker deps clean test coverage lint docker-local edgedb k8s-up k8s-down docker-delete docs update-local deploy-local

all: build

deps: ## Get dependencies
{{"\t"}}go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

lint: deps ## Lint the files
{{"\t"}}go vet
{{"\t"}}gocyclo -over 10 -ignore "generated" ./

test: lint ## Run unittests
{{"\t"}}go test -v ./...

coverage: ## Create test coverage report
{{"\t"}}go test -cover ./...
{{"\t"}}go test ./... -coverprofile=cover.out && go tool cover -html=cover.out -o coverage.html

goreleaser: tidy ## Creates local multiarch releases with GoReleaser
{{"\t"}}goreleaser release --snapshot --rm-dist

tidy: ## Pull in dependencies
{{"\t"}}go mod tidy && go mod vendor

fmt: ## Format All files
{{"\t"}}go fmt ./...

{{ .Name }}ctl: ## Builds the binary on the current platform
{{"\t"}}go build -mod=vendor -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'" -o $(PROJECT_NAME)ctl

docs: ## Builds the cli documentation
{{"\t"}}mkdir -p docs
{{"\t"}}./{{ .Name }}ctl docs

schema: ## Generates boilerplate code from the graph/schema.graphqls file
{{"\t"}}go run github.com/99designs/gqlgen update

build-tester:
{{"\t"}}docker build --target tester --build-arg VERSION=${VERSION} -t ${TEST_IMAGE} .

ci-lint: build-tester
{{"\t"}}docker run --rm ${TEST_IMAGE} go vet
{{"\t"}}docker run --rm ${TEST_IMAGE} gocyclo -over 16 -ignore "generated" ./

ci-unit: build-tester
{{"\t"}}docker run --rm ${TEST_IMAGE} go test -v ./...

ci-cover: build-tester
{{"\t"}}mkdir -p ./output
{{"\t"}}docker run --rm -v ./output:/out ${TEST_IMAGE} go test -cover ./...
{{"\t"}}docker run --rm -v ./output:/out ${TEST_IMAGE} go test ./... -coverprofile=/out/cover.out
{{"\t"}}docker run --rm -v ./output:/out ${TEST_IMAGE} go tool cover -html=/out/cover.out -o /out/coverage.html

ci-test: ci-lint ci-unit ci-cover

ci-build:
{{"\t"}}docker build --build-arg VERSION=${VERSION} -t ${IMAGE} .

clean: ## Reset everything
{{"\t"}}docker run --rm -v ./output:/out alpine rm -rf /out/*
{{"\t"}}git clean -fd
{{"\t"}}git clean -fx
{{"\t"}}git reset --hard

help: ## Display this help screen
{{"\t"}}@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
`)
}

func Dockerfile() []byte {
	return []byte(`FROM golang:alpine as builder
WORKDIR /app
ENV IMAGE_TAG=dev
RUN apk update && apk upgrade && apk add --no-cache ca-certificates git
RUN update-ca-certificates
ADD . /app/
ARG VERSION
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -ldflags="-s -w -X '{{ .Module }}/cmd.Version=${VERSION}'" -installsuffix cgo -o {{ .Name }}ctl .

FROM builder AS tester
RUN go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

FROM scratch

COPY --from=builder /app/{{ .Name }}ctl .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["./{{ .Name }}ctl"]
`)
}

func GoReleaser() []byte {
	return []byte(`env:
  - IMAGE_TAG={{.Tag}}


project_name: [% .Name %]ctl

builds:
  - ldflags: "-extldflags= -w -X '[% .Module %]/cmd.Version={{.Tag}}'"
    flags:
      - -mod=vendor
    env:
      - "CGO_ENABLED=0"
      - "GO111MODULE=on"
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
source:
  enabled: true
`)
}

func TestWorkflow() []byte {
	return []byte(`name: test
on: 
  push:
    paths:
      - '**.go'
  workflow_call:
jobs:
  test:
    strategy:
      matrix:
        go-version: [ 1.22.x ]
        os: [ ubuntu-latest ]
    runs-on: ${{ matrix.os }}
    steps:
      - name: Install Go
        uses: actions/setup-go@v2
        with:
          go-version: ${{ matrix.go-version }}
      - name: Checkout code
        uses: actions/checkout@v2
      - name: Test
        run: make test
      - name: Coverage
        run: make coverage
      - name: store coverage
        uses: actions/upload-artifact@v4
        with:
          name: test-coverage
          path: ./coverage.html 
`)
}

func ReleaseWorkflow() []byte {
	return []byte(`name: Build application

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

permissions:
  id-token: write # This is required for AWS authentication
  contents: write # This is required for actions/checkout
  issues: write # to be able to comment on released issues
  pull-requests: write # to be able to comment on released pull requests

jobs:
  build:
    uses: CoverWhale/cw-internal-developer-platform/.github/workflows/build.yml@v1.4.0
    with:
      system: prime
      application: prime-{{ .Name }}
`)
}

func Gitignore() []byte {
	return []byte(`{{ .Name }}ctl
cwgotctl*
dist/
output/
`)
}

func Dockerignore() []byte {
	return []byte(`Makefile
Dockerfile
.git 
.gitignore
.README.md
`)
}

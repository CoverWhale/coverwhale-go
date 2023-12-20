package tpl

func Makefile() []byte {
	return []byte(`PROJECT_NAME := "{{ .Name }}"
PKG := "{{ .Module }}"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
VERSION := $(shell if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then git describe --exact-match --tags HEAD 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)"; else echo "dev"; fi)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPRIVATE=github.com/CoverWhale

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

{{ .Name }}ctl: ## Builds the binary on the current platform
{{"\t"}}go build -mod=vendor -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'" -o $(PROJECT_NAME)ctl

docs: ## Builds the cli documentation
{{"\t"}}mkdir -p docs
{{"\t"}}./{{ .Name }}ctl docs

{{ if not .DisableDeployment }}
docker-local: ## Builds the container image and pushes to the local k8s registry
{{"\t"}}docker build -t localhost:50000/{{ .Name }}:latest .
{{"\t"}}docker push localhost:50000/{{ .Name }}:latest

docker-delete: ## Deletes the local docker image
{{"\t"}}docker image rm localhost:50000/{{ .Name }}:latest

update-local: docker-local ## Builds the container image and pushes to registry, rolls out the new container into the cluster
{{"\t"}}kubectl rollout restart deployment/{{ .Name }}

deploy-local: k8s-up {{ if .EnableGraphql }}schema{{- end}} edgedb {{ if .EnableNats }}nats{{ end }} {{ .Name }}ctl docker-local ## Creates a local k8s cluster, builds a docker image of {{ .Name }}, and pushes to local registry
{{"\t"}}./{{ .Name }}ctl deploy manual {{ if .EnableNats }}--nats-urls "nats:4222"{{ end }} | kubectl apply -f -
{{"\t"}}kubectl wait pods -l app={{ .Name }} --for condition=Ready --timeout=30s

generate-yaml: {{ .Name }}ctl
{{"\t"}}mkdir -p deployments/{dev,prod}
{{"\t"}}./{{ .Name }}ctl deploy manual $(ACTION) --ingress-class $(INGRESS_CLASS) --ingress-annotations $(ANNOTATIONS) $(INGRESS_TLS) --namespace {{ .Namespace }} \
{{"\t"}}{{"\t"}}--registry  {{ .ContainerRegistry }} --service-name {{ .Name }}-$(ENVIRONMENT) {{ if .EnableNats }}--nats-urls {{ .NatsServers }} {{ end }} \
        --ingress-host $(INGRESS) --version=$(TAG)> deployments/$(ENVIRONMENT)/{{ .Name }}.yaml

generate-dev: {{ .Name }}ctl ## Generate dev environment yaml for Argo
{{"\t"}}mkdir -p deployments/dev
{{"\t"}}ENVIRONMENT=dev INGRESS=dev-{{ .Name }}.{{ .Domain }} INGRESS_CLASS=nginx INGRESS_TLS=--ingress-tls TAG=latest ANNOTATIONS="cert-manager.io/cluster-issuer"="letsencrypt-prod" make generate-yaml

generate-prod: {{ .Name }}ctl ## Generate prod environment yaml for Argo
{{"\t"}}mkdir -p deployments/prod
{{"\t"}}ENVIRONMENT=prod INGRESS_CLASS=nginx INGRESS_TLS=--ingress-tls ANNOTATIONS="cert-manager.io/cluster-issuer"="letsencrypt-prod" INGRESS={{ .Name }}.{{ .Domain }} TAG=$(VERSION) make generate-yaml

k8s-up: ## Creates a local kubernetes cluster with a registry
{{"\t"}}k3d registry create {{ .Name }}-registry --port 50000
{{"\t"}}k3d cluster create {{ .Name }} --registry-use k3d-{{ .Name }}-registry:50000 --servers 3 -p "8080:80@loadbalancer"

k8s-down: ## Destroys the k8s cluster and registry
{{"\t"}}k3d registry delete {{ .Name }}-registry
{{"\t"}}k3d cluster delete {{ .Name }}
{{ end -}}

schema: ## Generates boilerplate code from the graph/schema.graphqls file
{{"\t"}}go run github.com/99designs/gqlgen update

edgedb: ## Deploys edgedb into the cluster
{{"\t"}}kubectl apply -f infra/edgedb.yaml
{{"\t"}}# kubectl wait pods -l app=edgedb --for condition=Ready --timeout=60s
{{"\t"}}@echo "EdgeDB UI: http://edgedb.127.0.0.1.nip.io:8080/ui?authToken=eyJhbGciOiJFQ0RILUVTIiwiZW5jIjoiQTI1NkdDTSIsImVwayI6eyJjcnYiOiJQLTI1NiIsImt0eSI6IkVDIiwieCI6IlFwNndsOG9UVzM0dWtfVzFJX2d4ZUhfdkxjUjloRnU2Ti1aSUZDc08yQjQiLCJ5IjoiNmdEbEloVHcyNUJjLTYzZzNzdDMyb0lTb1VfV3EzZlF6QUdzS21wUDQtcyJ9fQ..L5WeRSfDxBKNmb3D.xUBmRYzVlrkH75u6i4NZMiDn1ssFsfPkiUNLSq0FrcSigZKE_u4sJBsMb0xYQ3Tq_AGjhoIttYa7hICMjlFcnD0w1CYbDDgK3TKCMcgS4m_W_SZMhYvMX-eAatww6X_y7jc9XAdWOtMV8Mi5Q1vV5gQTgYKbenZ2Lpr9P3UU4eop9kOqfQ_bIoZ5k0r13BEafvem30nER5hnGudJgvbDu3BtB0G4Ng.5boI8diCBzkk3O8Ce1ZEVg"

{{ if .EnableNats }}
nats: ## Deploys NATS into the cluster
{{"\t"}}helm repo add nats https://nats-io.github.io/k8s/helm/charts/
{{"\t"}}helm repo update
{{"\t"}}helm install nats nats/nats -f infra/nats.yaml
{{ end }}

clean: ## Remove previous build
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
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -ldflags="-s -w -X '{{ .Module }}/cmd.Version=$(printf $(git describe --tags | cut -d '-' -f 1)-$(git rev-parse --short HEAD))'" -installsuffix cgo -o {{ .Name }}ctl .


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
        go-version: [ 1.19.x ]
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
        uses: actions/upload-artifact@v2
        with:
          name: test-coverage
          path: ./coverage.html 
`)
}

func ReleaseWorkflow() []byte {
	return []byte(`name: release and deploy
on:
  push:
    branches:
      - main
    paths-ignore:
      - 'docs/**'
      - 'deployments/**'
  pull_request:
    types: [opened, reopened, edited]
    paths:
      - '**.go'

env:
  DOCKER_REPO: ${{ env.ECR_REGISTRY }}/[% .Name %]
permissions:
  id-token: write
  contents: read
jobs:
  test:
    uses: ./.github/workflows/test.yaml
  release:
    permissions:
      id-token: write
      contents: write
    runs-on: ubuntu-latest
    needs: [test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v2
        with:
          token: ${{ secrets.WORKFLOW_GIT_ACCESS_TOKEN }}
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v1
        with:
          role-to-assume: ${{ env.ROLE_ARN }}
          role-session-name: ${{ env.ROLE_NAME }}
          aws-region: ${{ env.AWS_REGION }}
      - uses: aws-actions/amazon-ecr-login@v1
      - name: Build branch and push to ecr
        run: |
          docker build -t $DOCKER_REPO:${{github.sha}} -t $DOCKER_REPO:latest .
          docker push -a $DOCKER_REPO
        #use generated tag instead of latest
      - run: |
          make generate-dev TAG=${{github.sha}}
          make docs
      - name: create manifests and update docs
        if: ${{ github.event_name == 'push' }}
        run: |
          git config --global user.name "${{ github.event.repository.name }} CI"
          git config --global user.email "${{ env.USER_EMAIL }}"
          git add docs deployments
          git commit -m "update docs and manifests"
          git push
`)
}

func TaggedReleaseWorkflow() []byte {
	return []byte(`name: release and deploy
on:
  push:
    tags:
      - '*'
env:
  DOCKER_REPO: ${{ env.ECR_REGISTRY }}/[% .Name %]
permissions:
  id-token: write
  contents: read
jobs:
  test:
    uses: ./.github/workflows/test.yaml
  release:
    permissions:
      id-token: write
      contents: write
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout code
        uses: actions/checkout@v2 
        with:
          token: ${{ secrets.WORKFLOW_GIT_ACCESS_TOKEN }}
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v2
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.WORKFLOW_GIT_ACCESS_TOKEN }}
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v1
        with:
          role-to-assume: ${{ env.ROLE_ARN }}
          role-session-name: ${{ env.ROLE_NAME }}
          aws-region: ${{ env.AWS_REGION }}
      - uses: aws-actions/amazon-ecr-login@v1
      - name: Build with version and latest and push to ECR
        run: |
          docker build -t $DOCKER_REPO:${{ github.ref_name }} -t $DOCKER_REPO:latest . 
          docker push -a $DOCKER_REPO
      - run: |
          make generate-prod TAG=${{ github.ref_name }}
          make docs
      - name: create manifests and update docs
        run: |
          git config --global user.name "${{ github.event.repository.name }} CI"
          git config --global user.email "${{ env.USER_EMAIL }}"
          git add docs deployments
          git commit -m "update docs and manifests"
          git push origin HEAD:main
`)
}

func Gitignore() []byte {
	return []byte(`{{ .Name }}ctl*
cwgotctl*
dist/
`)
}

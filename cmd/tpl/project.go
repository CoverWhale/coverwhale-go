package tpl

func Server() []byte {
	return []byte(`package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "subcommand to control the server",
}

func init() {
    rootCmd.AddCommand(serverCmd)
    serverCmd.PersistentFlags().IntP("port", "p", 8080, "Server port")
    viper.BindPFlag("port", serverCmd.PersistentFlags().Lookup("port"))
}
`)
}

func ServerPackage() []byte {
	return []byte(`package server

import (
    {{ if not .DisableTelemetry -}}
    "context"
    {{- end }}
    "fmt"
    "math/rand"
    "net/http"
    "time"
    
    "github.com/CoverWhale/coverwhale-go/logging"
    cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
    {{ if not .DisableTelemetry -}}
    "github.com/CoverWhale/coverwhale-go/metrics"
    "go.opentelemetry.io/otel/attribute"
    {{- end }}
)

func GetRoutes(l *logging.Logger) []cwhttp.Route {
    return []cwhttp.Route{
        {
            Method: http.MethodGet,
            Path:   "/testing",
            Handler: &cwhttp.ErrHandler{
                Handler: testing,
                Logger:  l,
            },
        },
    }
}

{{ if not .DisableTelemetry -}}
func doMore(ctx context.Context) {
    // create new span from context
    _, span := metrics.NewTracer(ctx, "more sleepy")
    defer span.End()
    
    time.Sleep(500 * time.Millisecond)
}
{{- end }}

func testing(w http.ResponseWriter, r *http.Request) error {
    ie := r.Header.Get("internal-error")
    ce := r.Header.Get("client-error")
    
    if ie != "" {
        return fmt.Errorf("this is an internal error")
    }
    
    if ce != "" {
        return cwhttp.NewClientError(fmt.Errorf("uh oh spaghettios"), 400)
    }
    
    {{ if not .DisableTelemetry -}}
    // get new span
    ctx, span := metrics.NewTracer(r.Context(), "sleepytime")
    
    // if wanted define attributes for span
    attrs := []attribute.KeyValue{
        attribute.String("test", "this"),
    }
    span.SetAttributes(attrs...)
    defer span.End()
    {{- end }}
    
    rand.Seed(time.Now().UnixNano())
    i := rand.Intn(400-90+1) + 90
    
    sleep := time.Duration(i) * time.Millisecond
    time.Sleep(sleep)
    
    {{ if not .DisableTelemetry -}}
    // fake call to somethign that takes a long time
    doMore(ctx)
    {{- end }}
    
    resp := fmt.Sprintf("this works and took %dms\n", sleep.Milliseconds())
    
    w.Write([]byte(resp))
    return nil
}

func ExampleMiddleware(l *logging.Logger) func(h http.Handler) http.Handler {
    return func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Header.Get("Authorization") == "" {
                l.Info("unauthorized")
                w.WriteHeader(401)
                w.Write([]byte("unauthorized"))
                return
            }

            l.Info("in middleware")
            h.ServeHTTP(w, r)
        })
    }
}
`)
}

func ServerStart() []byte {
	return []byte(`package cmd 

import (
    "context"

    "{{ .Module }}/server"

    cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    {{ if not .DisableTelemetry -}}
    "github.com/CoverWhale/coverwhale-go/metrics"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    {{- end }}

)

var startCmd = &cobra.Command{
	Use:          "start",
	Short:        "starts the server",
	RunE:         start,
	SilenceUsage: true,
}

func init() {
	// attach start subcommand to server subcommand
	serverCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string ) error {
    ctx := context.Background()

    {{ if not .DisableTelemetry -}}
    // create new metrics exporter
    exp, err := metrics.NewOTLPExporter(ctx, {{ .MetricsUrl }}, otlptracehttp.WithInsecure())
    if err != nil {
        return err
    }

    // create global tracer provider
    tp, err := metrics.RegisterGlobalOTLPProvider(exp, "{{ .Name }}", Version)
    if err != nil {
        return err
    }
    {{- end }}

    s := cwhttp.NewHTTPServer(
        cwhttp.SetServerPort(viper.GetInt("port")),
    {{ if not .DisableTelemetry -}}
        cwhttp.SetTracerProvider(tp),
    {{- end }}
    )

    s.RegisterSubRouter("/api/v1", server.GetRoutes(s.Logger), server.ExampleMiddleware(s.Logger))

    errChan := make(chan error, 1)
    go s.Serve(errChan)
    s.AutoHandleErrors(ctx, errChan)

    return nil
} 
`)
}

func Main() []byte {
	return []byte(`package main

import "{{ .Module }}/cmd"

func main() {
        cmd.Execute()
}
`)
}

func Root() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
package cmd

import (
    "os"
    "strings"
    
    "github.com/CoverWhale/coverwhale-go/logging"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string
var cfg Config

var rootCmd = &cobra.Command{
    Use:   "{{ .Name }}",
    Short: "The app description",
}
var replacer = strings.NewReplacer("-", "_")

type Config struct {
    Port    int   {{ $tick }}mapstructure:"port"{{ $tick }}
}


func Execute() {
    viper.SetDefault("service-name", "{{ .Name }}-local")
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.{{ .Name }}.json)")
    rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {

    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)
        
        viper.AddConfigPath(home)
        viper.SetConfigType("json")
        viper.SetConfigName(".{{ .Name }}")
    }
    
    viper.SetEnvPrefix("{{ .Name }}")
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(replacer)
    
    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        logging.Debugf("using config %s", viper.ConfigFileUsed())
    }
    
    if err := viper.Unmarshal(&cfg); err != nil {
        cobra.CheckErr(err)
    }
}
`)
}

func Version() []byte {
	return []byte(`package cmd

import (
    "fmt"
    
    "github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Prints the version",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(Version)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
`)
}

func Deploy() []byte {
	return []byte(`package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Create k8s deployment info",
}

func init() {
    rootCmd.AddCommand(deployCmd)
    deployCmd.PersistentFlags().String("name", "{{ .Name }}", "Name of the app")
    viper.BindPFlag("name", deployCmd.PersistentFlags().Lookup("name"))
    deployCmd.PersistentFlags().String("registry", "k3d-{{ .Name }}-registry:50000", "Container registry")
    viper.BindPFlag("registry", deployCmd.PersistentFlags().Lookup("registry"))
    deployCmd.PersistentFlags().String("namespace", "default", "Deployment Namespace")
    viper.BindPFlag("namespace", deployCmd.PersistentFlags().Lookup("namespace"))
    deployCmd.PersistentFlags().String("version", "latest", "Container version (tag)")
    viper.BindPFlag("version", deployCmd.PersistentFlags().Lookup("version"))
    deployCmd.PersistentFlags().Int("service-port", 80, "k8s service port")
    viper.BindPFlag("service-port", deployCmd.PersistentFlags().Lookup("service-port"))
    deployCmd.PersistentFlags().String("ingress-host", "{{ .Name }}.127.0.0.1.nip.io", "k8s ingresss host")
    viper.BindPFlag("ingress-host", deployCmd.PersistentFlags().Lookup("ingress-host"))
    deployCmd.PersistentFlags().Bool("ingress-tls", false, "k8s ingresss tls")
    viper.BindPFlag("ingress-tls", deployCmd.PersistentFlags().Lookup("ingress-tls"))
    deployCmd.PersistentFlags().String("ingress-class", "", "k8s ingresss class name")
    viper.BindPFlag("ingress-class", deployCmd.PersistentFlags().Lookup("ingress-class"))
    deployCmd.PersistentFlags().Bool("insecure", false, "local insecure deployment")
    viper.BindPFlag("insecure", deployCmd.PersistentFlags().Lookup("insecure"))
    deployCmd.PersistentFlags().StringToString("ingress-annotations", map[string]string{}, "Annotations for the ingress")
	viper.BindPFlag("ingress-annotations", deployCmd.PersistentFlags().Lookup("ingress-annotations"))
}
`)
}

func Manual() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
package cmd

import (
    "fmt"
    
    "github.com/CoverWhale/coverwhale-go/k8s"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var manualCmd = &cobra.Command{
    Use:   "manual",
    Short: "Get manual k8s deployment",
    RunE:  manual,
}

func init() {
    deployCmd.AddCommand(manualCmd)
}

// manual creates the manual yaml
func manual(cmd *cobra.Command, args []string) error {

    dep, err := printDeployment()
    if err != nil {
        return err
    }
    
    service, err := printService()
    if err != nil {
        return err
    }
    
    ingress, err := printIngress()
    if err != nil {
        return err
    }
    
    secret, err := printSecret()
    if err != nil {
        return err
    }
    
    fmt.Printf("%s%s%s%s", dep, service, ingress, secret)
    
    return nil
}

func printDeployment() (string, error) {
    image := fmt.Sprintf("%s/%s:%s", viper.GetString("registry"), viper.GetString("name"), viper.GetString("version"))
    
    probe := k8s.HTTPProbe{
        Path:          "/healthz",
        Port:          viper.GetInt("port"),
        PeriodSeconds: 10,
        InitialDelay:  10,
    }
    
    c := k8s.NewContainer(viper.GetString("name"),
        k8s.ContainerImage(image),
        k8s.ContainerPort("http", viper.GetInt("port")),
        k8s.ContainerArgs([]string{"server", "start"}),
        // this needs set because K8s will create an environment variable in the pod with the name of the service underscore "port". This overrides that.
        k8s.ContainerEnvVar("{{ .Name | ToUpper }}_PORT", fmt.Sprintf("%d", viper.GetInt("port"))),
        k8s.ContainerLivenessProbeHTTP(probe),
    )
    
    p := k8s.NewPodSpec("{{ .Name }}",
        k8s.PodLabel("app", viper.GetString("name")),
        k8s.PodContainer(c),
    )
    
    d := k8s.NewDeployment("{{ .Name }}",
        k8s.DeploymentNamespace(viper.GetString("namespace")),
        k8s.DeploymentSelector("app", viper.GetString("name")),
        k8s.DeploymentPodSpec(p),
    )
    
    return k8s.MarshalYaml(d)

}

func printService() (string, error) {
    service := k8s.NewService(viper.GetString("name"),
        k8s.ServiceNamespace(viper.GetString("namespace")),
        k8s.ServicePort(viper.GetInt("service-port"), viper.GetInt("port")),
        k8s.ServiceSelector("app", viper.GetString("name")),
    )
    
    return k8s.MarshalYaml(service)
}

func printIngress() (string, error) {
    r := k8s.Rule{
        Host: viper.GetString("ingress-host"),
        TLS:  viper.GetBool("ingress-tls"),
        Paths: []k8s.Path{
            {
                Name:    "/",
                Service: viper.GetString("name"),
                Port:    viper.GetInt("service-port"),
                Type:    "Prefix",
            },
        },
    }
    
    ingress := k8s.NewIngress(viper.GetString("name"),
        k8s.IngressNamespace(viper.GetString("namespace")),
        k8s.IngressRule(r),
    )
    
    ingress.Annotations = map[string]string{
        "external-dns.alpha.kubernetes.io/hostname": viper.GetString("ingress-host"),
        "alb.ingress.kubernetes.io/listen-ports":    fmt.Sprintf({{ $tick }}[{"HTTP":%d,"HTTPS": 443}]{{ $tick }}, viper.GetInt("service-port")),
    }
    
    for k, v := range viper.GetStringMapString("ingress-annotations") {
        ingress.Annotations[k] = v
    }
    
    if viper.GetBool("ingress-tls") {
        f := k8s.IngressClass(viper.GetString("ingress-class"))
        f(&ingress)
    }
    
    return k8s.MarshalYaml(ingress)
}

func printSecret() (string, error) {
    if viper.GetString("secret-key") == "" {
        return "", nil
    }
    
    secret := k8s.NewSecret(viper.GetString("secret-name"),
        k8s.SecretData("apiKey", []byte(viper.GetString("secret-key"))),
        k8s.SecretNamespace(viper.GetString("namespace")),
    )
    
    return k8s.MarshalYaml(secret)
}
`)
}

func Makefile() []byte {
	return []byte(`PROJECT_NAME := "{{ .Name }}"
PKG := "{{ .Module }}"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
VERSION := $$(git describe HEAD)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPRIVATE=github.com/CoverWhale

.PHONY: all build docker deps clean test coverage lint docker-local k8s-up k8s-down docker-delete docs update-local deploy-local

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

{{ .Name }}ctl: tidy ## Builds the binary on the current platform
{{"\t"}}go build -mod=vendor -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'" -o $(PROJECT_NAME)ctl

docs: ## Builds the cli documentation
{{"\t"}}./{{ .Name }}ctl docs

{{ if not .DisableDeployment }}
docker-local: ## Builds the container image and pushes to the local k8s registry
{{"\t"}}docker build -t localhost:50000/{{ .Name }}:latest .
{{"\t"}}docker push localhost:50000/{{ .Name }}:latest

docker-delete: ## Deletes the local docker image
{{"\t"}}docker image rm localhost:50000/{{ .Name }}:latest

update-local: docker-local ## Builds the container image and pushes to registry, rolls out the new container into the cluster
{{"\t"}}kubectl rollout restart deployment/{{ .Name }}

deploy-local: k8s-up {{ .Name }}ctl docker-local ## Creates a local k8s cluster, builds a docker image of {{ .Name }}, and pushes to local registry
{{"\t"}}./{{ .Name }}ctl deploy manual | kubectl apply -f -
{{"\t"}}kubectl wait pods -l app={{ .Name }} --for condition=Ready --timeout=30s

generate-yaml: {{ .Name }}ctl
{{"\t"}}mkdir -p deployments/{dev,prod}
{{"\t"}}./{{ .Name }}ctl deploy manual $(ACTION) --ingress-class alb --ingress-annotations $(ANNOTATIONS) --ingress-tls --namespace prime \
        --registry  005364446802.dkr.ecr.us-east-1.amazonaws.com --service-name prime-{{ .Name }}-$(ENVIRONMENT) \
        --ingress-host $(INGRESS) --version=$(TAG)> deployments/$(ENVIRONMENT)/{{ .Name }}.yaml

generate-dev: {{ .Name }}ctl ## Generate dev environment yaml for Argo
{{"\t"}}ENVIRONMENT=dev INGRESS=dev-{{ .Name }}.prime.coverwhale.dev TAG=latest ANNOTATIONS="alb.ingress.kubernetes.io/group.name"="dev-apps-internal","alb.ingress.kubernetes.io/scheme"="internal","alb.ingress.kubernetes.io/target-type"="ip","alb.ingress.kubernetes.io/certificate-arn"="arn:aws:acm:us-east-1:005364446802:certificate/6e4aca2c-7087-4625-8ee3-49c8dfc29f5b" make generate-yaml

generate-prod: {{ .Name }}ctl ## Generate prod environment yaml for Argo
{{"\t"}}ENVIRONMENT=prod INGRESS={{ .Name }}.prime.coverwhale.com TAG=$(VERSION) make generate-yaml

k8s-up: ## Creates a local kubernetes cluster with a registry
{{"\t"}}k3d registry create {{ .Name }}-registry --port 50000
{{"\t"}}k3d cluster create {{ .Name }} --registry-use k3d-{{ .Name }}-registry:50000 --servers 3 -p "8080:80@loadbalancer"

k8s-down: ## Destroys the k8s cluster and registry
{{"\t"}}k3d registry delete {{ .Name }}-registry
{{"\t"}}k3d cluster delete {{ .Name }}
{{ end -}}

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
on: [push, workflow_call]
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
  DOCKER_REPO: "${{ secrets.ECR_REGISTRY }}"/[% .Name %]
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
          role-to-assume: arn:aws:iam::005364446802:role/GithubActionsPulumi
          role-session-name: github-actions-pulumi
          aws-region: ${{ secrets.AWS_REGION }}
      - uses: aws-actions/amazon-ecr-login@v1
      - name: Build branch and push to ecr
        run: |
          docker build -t $DOCKER_REPO:${{github.sha}}-t $DOCKER_REPO:latest .
          docker push -a $DOCKER_REPO
        #use generated tag instead of latest
      - run: make generate-dev TAG=${{github.sha}}
      - run: make docs
      - name: create manifests and update docs
        if: ${{ github.event_name == 'push' }}
        run: |
          git config --global user.name "${{ github.event.repository.name }} CI"
          git config --global user.email "automations@coverwhale.com"
          git add docs deployments
          git commit -m "update docs and manifests"
          git push
`)
}

func Gitignore() []byte {
	return []byte(`{{ .Name }}ctl*
cwgotctl*
`)
}

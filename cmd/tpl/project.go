// Copyright 2023 Cover Whale Insurance Solutions Inc.
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

func Docs() []byte {
	return []byte(`package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate cli documentation",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doc.GenMarkdownTree(rootCmd, "./docs")
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}

`)
}

func Service() []byte {
	return []byte(`package cmd

import (
    "github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
    Use:   "service",
    Short: "subcommand to control the service",
    // PersistentPostRun is used here because this is just a subcommand with no run function
    PersistentPreRun: bindServiceCmdFlags,
}

func init() {
    rootCmd.AddCommand(serviceCmd)
    serviceFlags(serviceCmd)
    natsFlags(serviceCmd)
}

func bindServiceCmdFlags(cmd *cobra.Command, args []string) {
    bindServiceFlags(cmd)
    bindNatsFlags(cmd)
}
`)
}

func ServicePackage() []byte {
	return []byte(`package service

import (
    {{ if .EnableTelemetry -}}
    "context"
    {{- end }}
    "fmt"
    "math/rand"
    "net/http"
    "time"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "github.com/CoverWhale/logr"
    cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
    {{ if .EnableTelemetry -}}
    "github.com/CoverWhale/coverwhale-go/metrics"
    "go.opentelemetry.io/otel/attribute"
    {{- end }}
)

func GetRoutes(l *logr.Logger) []cwhttp.Route {
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

func GetPlayground(srv *handler.Server) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodGet,
			Path:    "/playground",
			Handler: playground.Handler("GraphQL playground", "/api/v1/graphql/query"),
		},
	}
}

func GetApiQuery(srv *handler.Server) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodPost,
			Path:    "/query",
			Handler: srv,
		},
	}
}

{{ if not .EnableTelemetry -}}
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
    
    {{ if .EnableTelemetry -}}
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
    
    {{ if .EnableTelemetry -}}
    // fake call to somethign that takes a long time
    doMore(ctx)
    {{- end }}
    
    resp := fmt.Sprintf("this works and took %dms\n", sleep.Milliseconds())
    
    w.Write([]byte(resp))
    return nil
}

func ExampleMiddleware(l *logr.Logger) func(h http.Handler) http.Handler {
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

func ServiceStart() []byte {
	return []byte(`package cmd 

import (
    {{ if .EnableHTTP }}
    "context"

    cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
    {{ end }}
    "{{ .Module }}/service"
    "github.com/CoverWhale/logr"
    "github.com/invopop/jsonschema"
    "github.com/nats-io/nats.go/micro"
    "github.com/nats-io/nats.go"
    cwnats "github.com/CoverWhale/coverwhale-go/transports/nats"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    {{ if and .EnableHTTP .EnableTelemetry -}}"github.com/CoverWhale/coverwhale-go/metrics"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"{{- end }}
    {{ if .EnableGraphql }}"github.com/99designs/gqlgen/graphql/handler"
    "{{ .Module }}/graph"{{- end}}
)

var startCmd = &cobra.Command{
	Use:          "start",
	Short:        "starts the service",
	RunE:         start,
	SilenceUsage: true,
}

func init() {
	// attach start subcommand to service subcommand
	serviceCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string ) error {
    logger := logr.NewLogger()
    {{ if .EnableHTTP }}
    ctx := context.Background()

    {{ if .EnableTelemetry -}}
    // create new metrics exporter
    exp, err := metrics.NewOTLPExporter(ctx, "{{ .MetricsUrl }}", otlptracehttp.WithInsecure())
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
        {{ if .EnableTelemetry -}}
        cwhttp.SetTracerProvider(tp),
        {{- end }}
    )

    errChan := make(chan error, 1)
    {{- end }}

    {{ if .EnableGraphql }}resolver := &graph.Resolver{}{{- end }}

    config := micro.Config{
    	Name:        "{{ .Name }}",
    	Version:     "0.0.1",
    	Description: "An example application",
    }
    
    nc, err := nats.Connect(viper.GetString("nats_urls"))
    if err != nil {
    	logr.Fatal(err)
    }
    defer nc.Close()

    logger.SetOutput(cwnats.NewNatsLogger("prime.logs.{{ .Name }}", nc))
    
    svc, err := micro.AddService(nc, config)
    if err != nil {
    	logr.Fatal(err)
    }
    
    // add a singular handler as an endpoint
    svc.AddEndpoint("specific", cwnats.ErrorHandler(logger, service.SpecificHandler), micro.WithEndpointSubject("prime.services.{{ .Name }}.*.specific.get"))
    
    // add a handler group
    grp := svc.AddGroup("prime.services.{{ .Name }}.*.math", micro.WithGroupQueueGroup("{{ .Name }}"))
    grp.AddEndpoint("add",
    	cwnats.ErrorHandler(logger, service.Add),
    	micro.WithEndpointMetadata(map[string]string{
    		"description":     "adds two numbers",
    		"format":          "application/json",
    		"request_schema":  schemaString(&service.MathRequest{}),
    		"response_schema": schemaString(&service.MathResponse{}),
    	}),
	micro.WithEndpointSubject("add.get"),
    )
    grp.AddEndpoint("subtract",
    	cwnats.ErrorHandler(logger, service.Subtract),
    	micro.WithEndpointMetadata(map[string]string{
    		"description":     "subtracts two numbers",
    		"format":          "application/json",
    		"request_schema":  schemaString(&service.MathRequest{}),
    		"response_schema": schemaString(&service.MathResponse{}),
    	}),
	micro.WithEndpointSubject("subtract.get"),
    )
    
    {{ if not .EnableHTTP }}
    cwnats.HandleNotify(svc)
    {{- end }}

    {{ if .EnableHTTP }}
    service.Watch(n, "prime.{{ .Name }}.*")

    s.RegisterSubRouter("/api/v1", service.GetRoutes(s.Logger), service.ExampleMiddleware(s.Logger))
    {{ if .EnableGraphql }}
    srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
    s.RegisterSubRouter("/", service.GetPlayground(srv))
    s.RegisterSubRouter("/api/v1/graphql", service.GetApiQuery(srv))
    {{- end }}

    go s.Serve(errChan)
    s.AutoHandleErrors(ctx, errChan)
    {{- end }}

    return nil
} 

func schemaString(s any) string {
    schema := jsonschema.Reflect(s)
    data, err := schema.MarshalJSON()
    if err != nil {
    	logr.Fatal(err)
    }
    
    return string(data)
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
    
    "github.com/CoverWhale/logr"
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
    logger := logr.NewLogger()
    if err := viper.ReadInConfig(); err == nil {
        logger.Debugf("using config %s", viper.ConfigFileUsed())
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

func Flags() []byte {
	return []byte(`package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

//Flags are defined here. Because of the way Viper binds values, if the same flag name is called
// with viper.BindPFlag multiple times during init() the value will be overwritten. For example if
// two subcommands each have a flag called name but they each have their own default values,
// viper can overwrite any value passed in for one subcommand with the default value of the other subcommand.
// The answer here is to not use init() and instead use something like PersistentPreRun to bind the
// viper values. Using init for the cobra flags is ok, they are only in here to limit duplication of names.

// bindNatsFlags binds nats flag values to viper
func bindNatsFlags(cmd *cobra.Command) {
	viper.BindPFlag("nats_urls", cmd.Flags().Lookup("nats-urls"))
}

// natsFlags adds the nats flags to the passed in cobra command
func natsFlags(cmd *cobra.Command) {
    cmd.PersistentFlags().String("nats-urls", "nats://localhost:4222", "NATS server URL(s)")
}

// bindServiceFlags binds the secret flag values to viper
func bindServiceFlags(cmd *cobra.Command) {
    viper.BindPFlag("port", cmd.Flags().Lookup("port"))
    viper.BindPFlag("enable_verification", cmd.Flags().Lookup("enable-verification"))
    viper.BindPFlag("tempo_url", cmd.Flags().Lookup("tempo-url"))
}

// sererFlags adds the service flags to the passed in command
func serviceFlags(cmd *cobra.Command) {
    cmd.PersistentFlags().IntP("port", "p", 8080, "Server port")
    cmd.PersistentFlags().Bool("enable-verification", false, "Enable driver verification")
    cmd.PersistentFlags().String("tempo-url", "", "URL for Tempo")
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
    PersistentPreRun: bindDeployCmdFlags,
}

func init() {
    rootCmd.AddCommand(deployCmd)
    deployCmd.PersistentFlags().String("name", "{{ .Name }}", "Name of the app")
    deployCmd.PersistentFlags().String("registry", "k3d-{{ .Name }}-registry:50000", "Container registry")
    deployCmd.PersistentFlags().String("namespace", "default", "Deployment Namespace")
    deployCmd.PersistentFlags().String("version", "latest", "Container version (tag)")
    deployCmd.PersistentFlags().Int("service-port", 80, "k8s service port")
    deployCmd.PersistentFlags().String("service-name", "{{ .Name }}", "k8s service name")
    deployCmd.PersistentFlags().String("ingress-host", "{{ .Name }}.127.0.0.1.nip.io", "k8s ingresss host")
    deployCmd.PersistentFlags().Bool("ingress-tls", false, "k8s ingresss tls")
    deployCmd.PersistentFlags().String("ingress-class", "", "k8s ingresss class name")
    deployCmd.PersistentFlags().StringToString("ingress-annotations", map[string]string{}, "Annotations for the ingress")
    natsFlags(deployCmd)
    serviceFlags(deployCmd)
}

func bindDeployCmdFlags(cmd *cobra.Command, args []string) {
    viper.BindPFlag("name", cmd.Flags().Lookup("name"))
    viper.BindPFlag("registry", cmd.Flags().Lookup("registry"))
    viper.BindPFlag("namespace", cmd.Flags().Lookup("namespace"))
    viper.BindPFlag("version", cmd.Flags().Lookup("version"))
    viper.BindPFlag("service_port", cmd.Flags().Lookup("service-port"))
    viper.BindPFlag("service_name", cmd.Flags().Lookup("service-name"))
    viper.BindPFlag("ingress_host", cmd.Flags().Lookup("ingress-host"))
    viper.BindPFlag("ingress_tls", cmd.Flags().Lookup("ingress-tls"))
    viper.BindPFlag("ingress_class", cmd.Flags().Lookup("ingress-class"))
    viper.BindPFlag("insecure", cmd.Flags().Lookup("insecure"))
    viper.BindPFlag("ingress_annotations", cmd.Flags().Lookup("ingress-annotations"))
    viper.BindPFlag("tempo_url", cmd.Flags().Lookup("tempo-url"))
    bindServiceFlags(cmd)
    bindNatsFlags(cmd)
}
`)
}

func Manual() []byte {
	return []byte(`{{ $tick := "` + "`" + `" -}}
package cmd

import (
    "fmt"
    "strings"
    
    "github.com/CoverWhale/kopts"
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
    
    probe := kopts.HTTPProbe{
        Path:          "/healthz",
        Port:          viper.GetInt("port"),
        PeriodSeconds: 10,
        InitialDelay:  10,
    }

    // k8s doesn't like dashes in env names
    replacer := strings.NewReplacer("-", "_")    
    c := kopts.NewContainer(viper.GetString("name"),
        kopts.ContainerImage(image),
        kopts.ContainerPort("http", viper.GetInt("port")),
        kopts.ContainerArgs([]string{"service", "start"}),
        // this needs set because K8s will create an environment variable in the pod with the name of the service underscore "port". This overrides that.
        kopts.ContainerEnvVar(replacer.Replace("{{ .Name | ToUpper }}_PORT"), fmt.Sprintf("%d", viper.GetInt("port"))),
        kopts.ContainerLivenessProbeHTTP(probe),
        kopts.ContainerEnvVar("{{ .Name | ToUpper }}_NATS_URLS", viper.GetString("nats_urls")),
    )
    
    p := kopts.NewPodSpec("{{ .Name }}",
        kopts.PodLabel("app", viper.GetString("name")),
        kopts.PodContainer(c),
    )
    
    d := kopts.NewDeployment("{{ .Name }}",
        kopts.DeploymentNamespace(viper.GetString("namespace")),
        kopts.DeploymentSelector("app", viper.GetString("name")),
        kopts.DeploymentPodSpec(p),
    )
    
    return kopts.MarshalYaml(d)

}

func printService() (string, error) {
    service := kopts.NewService(viper.GetString("service_name"),
        kopts.ServiceNamespace(viper.GetString("namespace")),
        kopts.ServicePort(viper.GetInt("service_port"), viper.GetInt("port")),
        kopts.ServiceSelector("app", viper.GetString("name")),
    )
    
    return kopts.MarshalYaml(service)
}

func printIngress() (string, error) {
    r := kopts.Rule{
        Host: viper.GetString("ingress_host"),
        TLS:  viper.GetBool("ingress_tls"),
        Paths: []kopts.Path{
            {
                Name:    "/",
                Service: viper.GetString("name"),
                Port:    viper.GetInt("service_port"),
                Type:    "Prefix",
            },
        },
    }
    
    ingress := kopts.NewIngress(viper.GetString("name"),
        kopts.IngressNamespace(viper.GetString("namespace")),
        kopts.IngressRule(r),
    )
    
    ingress.Annotations = map[string]string{
        "external-dns.alpha.kubernetes.io/hostname": viper.GetString("ingress_host"),
        "alb.ingress.kubernetes.io/listen-ports":    fmt.Sprintf({{ $tick }}[{"HTTP":%d,"HTTPS": 443}]{{ $tick }}, viper.GetInt("service_port")),
    }
    
    for k, v := range viper.GetStringMapString("ingress_annotations") {
        ingress.Annotations[k] = v
    }
    
    if viper.GetBool("ingress_tls") {
        f := kopts.IngressClass(viper.GetString("ingress_class"))
        f(&ingress)
    }
    
    return kopts.MarshalYaml(ingress)
}

func printSecret() (string, error) {
    if viper.GetString("secret_key") == "" {
        return "", nil
    }
    
    secret := kopts.NewSecret(viper.GetString("secret_name"),
        kopts.SecretData("apiKey", []byte(viper.GetString("secret_key"))),
        kopts.SecretNamespace(viper.GetString("namespace")),
    )
    
    return kopts.MarshalYaml(secret)
}
`)
}

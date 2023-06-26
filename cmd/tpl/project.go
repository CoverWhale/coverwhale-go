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

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
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
    {{ if .EnableGraphql }}
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/CoverWhale/{{ .Name }}/graph"
    {{- end}}

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
    {{ if not .DisableTelemetry -}}
        cwhttp.SetTracerProvider(tp),
    {{- end }}
    )

    {{ if .EnableNats }}
    backend := server.NewNatsBackend("{{ .NatsServers }}")
    if err := backend.Connect(); err != nil {
        return err
    }

    backend.Watch("{{ .NatsSubject }}")
    {{ end }}

    s.RegisterSubRouter("/api/v1", server.GetRoutes(s.Logger), server.ExampleMiddleware(s.Logger))
    {{ if .EnableGraphql }}
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
    s.RegisterSubRouter("/", server.GetPlayground(srv))
    s.RegisterSubRouter("/api/v1/graphql", server.GetApiQuery(srv))
    {{- end}}

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

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/CoverWhale/coverwhale-go/cmd/tpl"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Creates a new HTTP server template",
	Long:  `Creates a new HTTP server for a Cover Whale microservice`,
	RunE:  server,
}

func init() {
	newCmd.AddCommand(serverCmd)
	serverCmd.Flags().Bool("disable-telemetry", false, "Enable opentelemetry integration")
	viper.BindPFlag("server.disable_telemetry", serverCmd.Flags().Lookup("disable-telemetry"))
	serverCmd.Flags().StringP("name", "n", "", "Application name")
	serverCmd.MarkFlagRequired("name")
	viper.BindPFlag("server.name", serverCmd.Flags().Lookup("name"))
	serverCmd.Flags().Bool("disable-deployment", false, "Disables Kubernetes deployment generation")
	viper.BindPFlag("server.disable_deployment", serverCmd.Flags().Lookup("disable-deployment"))
	serverCmd.PersistentFlags().String("metrics-url", "localhost:4318", "Endpoint for metrics exporter")
	viper.BindPFlag("server.metrics_url", serverCmd.PersistentFlags().Lookup("metrics-url"))
	serverCmd.PersistentFlags().Bool("enable-nats", false, "Enables NATS integration")
	viper.BindPFlag("server.enable_nats", serverCmd.PersistentFlags().Lookup("enable-nats"))
	serverCmd.PersistentFlags().String("nats-subject", "", "Subject(s) to listen on")
	viper.BindPFlag("server.nats_subject", serverCmd.PersistentFlags().Lookup("nats-subject"))
	serverCmd.PersistentFlags().String("nats-servers", "localhost:4222", "Nats server URLs")
	viper.BindPFlag("server.nats_servers", serverCmd.PersistentFlags().Lookup("nats-servers"))
	serverCmd.PersistentFlags().Bool("enable-graphql", false, "Enables GraphQL integration")
	viper.BindPFlag("server.enable_graphql", serverCmd.PersistentFlags().Lookup("enable-graphql"))
}

type Delims struct {
	First  string
	Second string
}

type CreateFileFromTemplate func(s *Server) error

var dd Delims

var opts []CreateFileFromTemplate

func server(cmd *cobra.Command, args []string) error {
	mod := modInfo()
	if mod == "command-line-arguments" {
		return fmt.Errorf("you must initialize a module with `go mod init <MODNAME>`")
	}
	cfg.Server.Module = mod

	if !cfg.Debug {
		dirs := []string{"./cmd", "./server", "./graph", "./dbschema", "./.github/workflows", "./infra"}
		for _, v := range dirs {
			if _, err := os.Stat(v); os.IsNotExist(err) {
				if err := os.MkdirAll(v, 0755); err != nil {
					log.Printf("error creating path: %s", err)
					os.Exit(1)
				}
			}
		}
	}

	// files we always create
	opts = []CreateFileFromTemplate{
		createMain(dd),
		createRoot(dd),
		createServer(dd),
		createServerStart(dd),
		createServerPackage(dd),
		createVersion(dd),
		createMakefile(dd),
		createDockerfile(dd),
		createGoReleaser(Delims{First: "[%", Second: "%]"}),
		createTestWorkflow(Delims{First: "[%", Second: "%]"}),
		createReleaseWorkflow(Delims{First: "[%", Second: "%]"}),
		createGitignore(dd),
		createEdgedbToml(dd),
		createEdgedbDefault(dd),
		createEdgedbFuture(dd),
		createEdgeDBInfra(dd),
	}

	// deployment
	if !cfg.Server.DisableDeployment {
		opts = append(opts,
			createDeploy(dd),
			createManual(dd),
		)
	}

	// graphql
	if cfg.Server.EnableGraphql {
		opts = append(opts,
			createClient(dd),
			createGQLGen(dd),
			createSchemaGraphql(dd),
			createResolver(dd),
			createModelsGen(dd),
			createSchemaResolver(dd),
			createTools(dd),
		)
	}

	// nats
	if cfg.Server.EnableNats {
		opts = append(opts, createNats(dd))
	}

	err := cfg.Server.CreateFilesFromTemplates(opts...)

	if err != nil {
		return err
	}

	return nil
}

// core files needed for any project
// templates under project.go
func createMain(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("main.go", tpl.Main(), dd)
	}
}

func createRoot(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/root.go", tpl.Root(), dd)
	}
}

func createServer(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/server.go", tpl.Server(), dd)
	}
}

func createServerStart(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/start.go", tpl.ServerStart(), dd)
	}
}

func createServerPackage(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("server/server.go", tpl.ServerPackage(), dd)
	}
}

func createVersion(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/version.go", tpl.Version(), dd)
	}
}

// only if deployment is enabled
func createDeploy(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/deploy.go", tpl.Deploy(), dd)
	}
}

func createManual(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("cmd/manual.go", tpl.Manual(), dd)
	}
}

// build and deployments
// templates under deployment.go
func createMakefile(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("Makefile", tpl.Makefile(), dd)
	}
}

func createDockerfile(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("Dockerfile", tpl.Dockerfile(), dd)
	}
}

func createGoReleaser(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile(".goreleaser.yaml", tpl.GoReleaser(), dd)
	}
}

func createTestWorkflow(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile(".github/workflows/test.yaml", tpl.TestWorkflow(), dd)
	}
}

func createReleaseWorkflow(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile(".github/workflows/release.yaml", tpl.ReleaseWorkflow(), dd)
	}
}

func createGitignore(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile(".gitignore", tpl.Gitignore(), dd)
	}
}

// graphql stuff
// templates under graphql.go
func createClient(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("graph/client.go", tpl.Client(), dd)
	}
}

func createGQLGen(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("gqlgen.yaml", tpl.GQLGen(), dd)
	}
}

func createSchemaGraphql(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("graph/schema.graphqls", tpl.SchemaGraphqls(), dd)
	}
}

func createResolver(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("graph/resolver.go", tpl.Resolvers(), dd)
	}
}

func createModelsGen(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("graph/models_gen.go", tpl.ModelsGen(), dd)
	}
}

func createSchemaResolver(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("graph/schema.resolvers.go", tpl.SchemaResolvers(), dd)
	}
}

func createTools(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("tools.go", tpl.Tools(), dd)
	}
}

// nats
// templates under nats.go
func createNats(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("server/nats.go", tpl.Nats(), dd)
	}
}

// edgedb
// templates under edgedb.go
func createEdgedbToml(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("edgedb.toml", tpl.EdgeDBToml(), dd)
	}
}
func createEdgedbDefault(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("dbschema/default.esdl", tpl.DefaultEsdl(), dd)
	}
}
func createEdgedbFuture(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("dbschema/future.esdl", tpl.FutureEsdl(), dd)
	}
}
func createEdgeDBInfra(dd Delims) CreateFileFromTemplate {
	return func(s *Server) error {
		return cfg.Server.createOrPrintFile("infra/edgedb.yaml", tpl.EdgeDBInfra(), dd)
	}
}

func (s *Server) CreateFilesFromTemplates(opts ...CreateFileFromTemplate) error {
	for _, t := range opts {
		if err := t(s); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) createOrPrintFile(n string, b []byte, d Delims) error {
	if d.First == "" && d.Second == "" {
		d.First = "{{"
		d.Second = "}}"
	}

	if cfg.Debug {
		return s.handleOutput(os.Stdout, b, d)
	}

	f, err := os.Create(n)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}

	defer f.Close()

	return s.handleOutput(f, b, d)
}

func (s *Server) handleOutput(w io.Writer, b []byte, d Delims) error {
	fmap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	temp := template.Must(template.New("file").Delims(d.First, d.Second).Funcs(fmap).Parse(string(b)))
	if err := temp.Execute(w, s); err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	return nil
}

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

var dd Delims

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

	if err := cfg.Server.createMain(); err != nil {
		return err
	}

	if err := cfg.Server.createRoot(); err != nil {
		return err
	}

	if err := cfg.Server.createServer(); err != nil {
		return err
	}

	if err := cfg.Server.createServerStart(); err != nil {
		return err
	}

	if err := cfg.Server.createVersion(); err != nil {
		return err
	}

	if err := cfg.Server.createServerPackage(); err != nil {
		return err
	}

	if !cfg.Server.DisableDeployment {
		if err := cfg.Server.createDeploy(); err != nil {
			return err
		}

		if err := cfg.Server.createManual(); err != nil {
			return err
		}
	}

	if err := cfg.Server.createMakefile(); err != nil {
		return err
	}

	if err := cfg.Server.createDockerfile(); err != nil {
		return err
	}

	if err := cfg.Server.createGoReleaser(); err != nil {
		return err
	}

	if err := cfg.Server.createTestWorkflow(); err != nil {
		return err
	}

	if err := cfg.Server.createReleaseWorkflow(); err != nil {
		return err
	}

	if err := cfg.Server.createGitignore(); err != nil {
		return err
	}

	if cfg.Server.EnableNats {
		if err := cfg.Server.createNats(); err != nil {
			return err
		}
	}

	// all graphql stuff
	if cfg.Server.EnableGraphql {
		if err := cfg.Server.createClient(); err != nil {
			return err
		}

		if err := cfg.Server.createGQLGen(); err != nil {
			return err
		}

		if err := cfg.Server.createSchemaGraphql(); err != nil {
			return err
		}

		if err := cfg.Server.createTools(); err != nil {
			return err
		}
	}

	// edgedb
	if err := cfg.Server.createEdgedbToml(); err != nil {
		return err
	}

	if err := cfg.Server.createEdgedbDefault(); err != nil {
		return err
	}

	if err := cfg.Server.createEdgedbFuture(); err != nil {
		return err
	}

	if err := cfg.Server.createEdgeDBInfra(); err != nil {
		return err
	}

	return nil
}

// template under project.go
func (s *Server) createMain() error {
	return cfg.Server.createOrPrintFile("main.go", tpl.Main(), dd)
}

func (s *Server) createRoot() error {
	return cfg.Server.createOrPrintFile("cmd/root.go", tpl.Root(), dd)
}

func (s *Server) createServer() error {
	return cfg.Server.createOrPrintFile("cmd/server.go", tpl.Server(), dd)
}

func (s *Server) createServerStart() error {
	return cfg.Server.createOrPrintFile("cmd/start.go", tpl.ServerStart(), dd)
}

func (s *Server) createServerPackage() error {
	return cfg.Server.createOrPrintFile("server/server.go", tpl.ServerPackage(), dd)
}

func (s *Server) createVersion() error {
	return cfg.Server.createOrPrintFile("cmd/version.go", tpl.Version(), dd)
}

func (s *Server) createDeploy() error {
	return cfg.Server.createOrPrintFile("cmd/deploy.go", tpl.Deploy(), dd)
}

func (s *Server) createManual() error {
	return cfg.Server.createOrPrintFile("cmd/manual.go", tpl.Manual(), dd)
}

// template under graphql.go
func (s *Server) createClient() error {
	return cfg.Server.createOrPrintFile("graph/client.go", tpl.Client(), dd)
}

func (s *Server) createGQLGen() error {
	return cfg.Server.createOrPrintFile("gqlgen.yaml", tpl.GQLGen(), dd)
}

func (s *Server) createSchemaGraphql() error {
	return cfg.Server.createOrPrintFile("graph/schema.graphqls", tpl.SchemaGraphqls(), dd)
}

func (s *Server) createTools() error {
	return cfg.Server.createOrPrintFile("tools.go", tpl.Tools(), dd)
}

// template under deployment.go
func (s *Server) createMakefile() error {
	return cfg.Server.createOrPrintFile("Makefile", tpl.Makefile(), dd)
}

func (s *Server) createDockerfile() error {
	return cfg.Server.createOrPrintFile("Dockerfile", tpl.Dockerfile(), dd)
}

func (s *Server) createGoReleaser() error {
	return cfg.Server.createOrPrintFile(".goreleaser.yaml", tpl.GoReleaser(), Delims{First: "[%", Second: "%]"})
}

func (s *Server) createTestWorkflow() error {
	return cfg.Server.createOrPrintFile(".github/workflows/test.yaml", tpl.TestWorkflow(), Delims{First: "[%", Second: "%]"})
}

func (s *Server) createReleaseWorkflow() error {
	return cfg.Server.createOrPrintFile(".github/workflows/release.yaml", tpl.ReleaseWorkflow(), Delims{First: "[%", Second: "%]"})
}

func (s *Server) createGitignore() error {
	return cfg.Server.createOrPrintFile(".gitignore", tpl.Gitignore(), dd)
}

// template under nats.go
func (s *Server) createNats() error {
	return cfg.Server.createOrPrintFile("server/nats.go", tpl.Nats(), dd)
}

// template under edgedb.go
func (s *Server) createEdgedbToml() error {
	return cfg.Server.createOrPrintFile("edgedb.toml", tpl.EdgeDBToml(), dd)
}

func (s *Server) createEdgedbDefault() error {
	return cfg.Server.createOrPrintFile("dbschema/default.esdl", tpl.DefaultEsdl(), dd)
}

func (s *Server) createEdgedbFuture() error {
	return cfg.Server.createOrPrintFile("dbschema/future.esdl", tpl.FutureEsdl(), dd)
}

func (s *Server) createEdgeDBInfra() error {
	return cfg.Server.createOrPrintFile("infra/edgedb.yaml", tpl.EdgeDBInfra(), dd)
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

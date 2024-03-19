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

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Creates a new service",
	Long:  `Creates a new Cover Whale microservice from a template`,
	RunE:  service,
}

func init() {
	newCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().Bool("enable-telemetry", false, "Enable opentelemetry integration")
	viper.BindPFlag("service.enable_telemetry", serviceCmd.Flags().Lookup("enable-telemetry"))
	serviceCmd.Flags().StringP("name", "n", "", "Application name")
	serviceCmd.MarkFlagRequired("name")
	serviceCmd.PersistentFlags().String("namespace", "default", "Namespace for deployment")
	viper.BindPFlag("service.namespace", serviceCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("service.name", serviceCmd.Flags().Lookup("name"))
	serviceCmd.Flags().Bool("disable-deployment", false, "Disables Kubernetes deployment generation")
	viper.BindPFlag("service.disable_deployment", serviceCmd.Flags().Lookup("disable-deployment"))
	serviceCmd.PersistentFlags().String("metrics-url", "localhost:4318", "Endpoint for metrics exporter")
	viper.BindPFlag("service.metrics_url", serviceCmd.PersistentFlags().Lookup("metrics-url"))
	serviceCmd.PersistentFlags().Bool("enable-http", false, "Enables HTTP integration")
	viper.BindPFlag("service.enable_http", serviceCmd.PersistentFlags().Lookup("enable-http"))
	serviceCmd.PersistentFlags().String("nats-service", "", "NATS server urls")
	viper.BindPFlag("service.nats_servers", serviceCmd.PersistentFlags().Lookup("nats-servers"))
	serviceCmd.PersistentFlags().Bool("enable-graphql", false, "Enables GraphQL integration")
	viper.BindPFlag("service.enable_graphql", serviceCmd.PersistentFlags().Lookup("enable-graphql"))
	serviceCmd.PersistentFlags().String("container-registry", "example.com", "URL for container registry")
	viper.BindPFlag("service.container_registry", serviceCmd.PersistentFlags().Lookup("container-registry"))
	serviceCmd.PersistentFlags().String("domain", "example.com", "Domain for ingress URLs")
	viper.BindPFlag("service.domain", serviceCmd.PersistentFlags().Lookup("domain"))
}

type Delims struct {
	First  string
	Second string
}

type CreateFileFromTemplate func(s *Service) error

var dd Delims

func service(cmd *cobra.Command, args []string) error {
	mod := modInfo()
	if mod == "command-line-arguments" {
		return fmt.Errorf("you must initialize a module with `go mod init <MODNAME>`")
	}
	cfg.Service.Module = mod

	if !cfg.Debug {
		dirs := []string{"./cmd", "./service", "./graph", "./dbschema", "./.github/workflows", "./infra"}
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
	opts := []CreateFileFromTemplate{
		createMain(dd),
		createRoot(dd),
		createService(dd),
		createServiceStart(dd),
		createVersion(dd),
		createMakefile(dd),
		createDockerfile(dd),
		createGoReleaser(Delims{First: "[%", Second: "%]"}),
		createTestWorkflow(Delims{First: "[%", Second: "%]"}),
		createReleaseWorkflow(Delims{First: "[%", Second: "%]"}),
		createTaggedReleaseWorkflow(Delims{First: "[%", Second: "%]"}),
		createGitignore(dd),
		createEdgedbToml(dd),
		createEdgedbDefault(dd),
		createEdgeDBInfra(dd),
		createFlags(dd),
		createDocs(dd),
		createNats(dd),
		createNatsInfra(dd),
		createClientCmd(dd),
		createQueryCmd(dd),
		createNatsCmdHelper(dd),
	}

	if cfg.Service.EnableHTTP {
		opts = append(opts,
			createServicePackage(dd),
		)
	}

	// deployment
	if !cfg.Service.DisableDeployment {
		opts = append(opts,
			createDeploy(dd),
			createManual(dd),
		)
	}

	// graphql
	if cfg.Service.EnableGraphql {
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

	err := cfg.Service.CreateFilesFromTemplates(opts...)

	if err != nil {
		return err
	}

	return nil
}

// core files needed for any project
// templates under project.go
func createMain(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("main.go", tpl.Main(), dd)
	}
}

func createRoot(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/root.go", tpl.Root(), dd)
	}
}

func createService(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/service.go", tpl.Service(), dd)
	}
}

func createServiceStart(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/start.go", tpl.ServiceStart(), dd)
	}
}

func createServicePackage(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("service/server.go", tpl.ServicePackage(), dd)
	}
}

func createVersion(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/version.go", tpl.Version(), dd)
	}
}

func createFlags(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/flags.go", tpl.Flags(), dd)
	}
}

func createDocs(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/docs.go", tpl.Docs(), dd)
	}
}

func createNatsCmdHelper(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/nats.go", tpl.NatsHelper(), dd)
	}
}

// only if deployment is enabled
func createDeploy(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/deploy.go", tpl.Deploy(), dd)
	}
}

func createManual(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/manual.go", tpl.Manual(), dd)
	}
}

func createClientCmd(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/client.go", tpl.CmdClient(), dd)
	}
}

func createQueryCmd(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("cmd/query.go", tpl.Query(), dd)
	}
}

// build and deployments
// templates under deployment.go
func createMakefile(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("Makefile", tpl.Makefile(), dd)
	}
}

func createDockerfile(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("Dockerfile", tpl.Dockerfile(), dd)
	}
}

func createGoReleaser(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile(".goreleaser.yaml", tpl.GoReleaser(), dd)
	}
}

func createTestWorkflow(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile(".github/workflows/test.yaml", tpl.TestWorkflow(), dd)
	}
}

func createReleaseWorkflow(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile(".github/workflows/release.yaml", tpl.ReleaseWorkflow(), dd)
	}
}

func createTaggedReleaseWorkflow(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile(".github/workflows/tagged_release.yaml", tpl.TaggedReleaseWorkflow(), dd)
	}
}

func createGitignore(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile(".gitignore", tpl.Gitignore(), dd)
	}
}

// graphql stuff
// templates under graphql.go
func createClient(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("graph/client.go", tpl.Client(), dd)
	}
}

func createGQLGen(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("gqlgen.yaml", tpl.GQLGen(), dd)
	}
}

func createSchemaGraphql(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("graph/schema.graphqls", tpl.SchemaGraphqls(), dd)
	}
}

func createResolver(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("graph/resolver.go", tpl.Resolvers(), dd)
	}
}

func createModelsGen(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("graph/models_gen.go", tpl.ModelsGen(), dd)
	}
}

func createSchemaResolver(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("graph/schema.resolvers.go", tpl.SchemaResolvers(), dd)
	}
}

func createTools(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("tools.go", tpl.Tools(), dd)
	}
}

// nats
// templates under nats.go
func createNats(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("service/nats.go", tpl.Nats(), dd)
	}
}

// edgedb
// templates under edgedb.go
func createEdgedbToml(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("edgedb.toml", tpl.EdgeDBToml(), dd)
	}
}
func createEdgedbDefault(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("dbschema/default.esdl", tpl.DefaultEsdl(), dd)
	}
}
func createEdgeDBInfra(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("infra/edgedb.yaml", tpl.EdgeDBInfra(), dd)
	}
}

func createNatsInfra(dd Delims) CreateFileFromTemplate {
	return func(s *Service) error {
		return cfg.Service.createOrPrintFile("infra/nats.yaml", tpl.NATSInfra(), dd)
	}
}

func (s *Service) CreateFilesFromTemplates(opts ...CreateFileFromTemplate) error {
	for _, template := range opts {
		if err := template(s); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) createOrPrintFile(n string, b []byte, d Delims) error {
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

func (s *Service) handleOutput(w io.Writer, b []byte, d Delims) error {
	fmap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}
	temp := template.Must(template.New("file").Delims(d.First, d.Second).Funcs(fmap).Parse(string(b)))
	if err := temp.Execute(w, s); err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	return nil
}

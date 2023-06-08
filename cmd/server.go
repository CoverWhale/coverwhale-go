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
	serverCmd.Flags().StringP("name", "n", "coverwhale-app", "Application name")
	viper.BindPFlag("server.name", serverCmd.Flags().Lookup("name"))
	serverCmd.Flags().Bool("disable-deployment", false, "Disables Kubernetes deployment generation")
	viper.BindPFlag("server.disable_deployment", serverCmd.Flags().Lookup("disable-deployment"))
}

func server(cmd *cobra.Command, args []string) error {
	mod := modInfo()
	if mod == "command-line-arguments" {
		return fmt.Errorf("you must initialize a module with `go mod init <MODNAME>`")
	}
	cfg.Server.Module = mod

	if !cfg.Debug {
		dirs := []string{"./cmd", "./server"}
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

	return nil
}

func (s *Server) createMain() error {
	return cfg.Server.createOrPrintFile("main.go", tpl.Main())
}

func (s *Server) createRoot() error {
	return cfg.Server.createOrPrintFile("cmd/root.go", tpl.Root())
}

func (s *Server) createServer() error {
	return cfg.Server.createOrPrintFile("cmd/server.go", tpl.Server())
}

func (s *Server) createServerStart() error {
	return cfg.Server.createOrPrintFile("cmd/start.go", tpl.ServerStart())
}

func (s *Server) createServerPackage() error {
	return cfg.Server.createOrPrintFile("server/server.go", tpl.ServerPackage())
}

func (s *Server) createVersion() error {
	return cfg.Server.createOrPrintFile("cmd/version.go", tpl.Version())
}

func (s *Server) createDeploy() error {
	return cfg.Server.createOrPrintFile("cmd/deploy.go", tpl.Deploy())
}

func (s *Server) createManual() error {
	return cfg.Server.createOrPrintFile("cmd/manual.go", tpl.Manual())
}

func (s *Server) createMakefile() error {
	return cfg.Server.createOrPrintFile("Makefile", tpl.Makefile())
}

func (s *Server) createDockerfile() error {
	return cfg.Server.createOrPrintFile("Dockerfile", tpl.Dockerfile())
}

func (s *Server) createOrPrintFile(n string, b []byte) error {

	if cfg.Debug {
		return s.handleOutput(os.Stdout, b)
	}

	f, err := os.Create(n)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}

	defer f.Close()

	return s.handleOutput(f, b)
}

func (s *Server) handleOutput(w io.Writer, b []byte) error {
	fmap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	temp := template.Must(template.New("file").Funcs(fmap).Parse(string(b)))
	if err := temp.Execute(w, s); err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	return nil
}
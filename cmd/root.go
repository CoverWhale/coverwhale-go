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

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "sgoctl",
	Short: "Create an opinionated application",
}

// replace dash with underscore for environment variables
var replacer = strings.NewReplacer("-", "_")
var cfgFile string
var cfg Config

type Config struct {
	Debug   bool    `mapstructure:"debug"`
	Service Service `mapstructure:"service"`
	Level   slog.LevelVar
}
type Service struct {
	Name              string `mapstructure:"name"`
	Namespace         string `mapstructure:"namespace"`
	Domain            string `mapstructure:"domain"`
	Module            string
	EnableTelemetry   bool   `mapstructure:"enable_telemetry"`
	DisableDeployment bool   `mapstructure:"disable_deployment"`
	MetricsUrl        string `mapstructure:"metrics_url"`
	EnableHTTP        bool   `mapstructure:"enable_http"`
	NatsSubject       string `mapstructure:"nats_subject"`
	NatsServers       string `mapstructure:"nats_servers"`
	EnableGraphql     bool   `mapstructure:"enable_graphql"`
	EnableEdgeDB      bool   `mapstructure:"enable_edgedb"`
	ContainerRegistry string `mapstructure:"container_registry"`
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sgo.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Print output instead of creating files")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(".sgo")
	}

	viper.SetEnvPrefix("sgo")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := viper.ReadInConfig(); err == nil {
		logger.Debug(fmt.Sprintf("using config %s", viper.ConfigFileUsed()))
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		cobra.CheckErr(err)
	}

}

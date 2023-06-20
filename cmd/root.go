package cmd

import (
	"os"
	"strings"

	"github.com/CoverWhale/coverwhale-go/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "cwgo",
	Short: "Create an opinionated application",
}

// replace dash with underscore for environment variables
var replacer = strings.NewReplacer("-", "_")
var cfgFile string
var cfg Config

type Config struct {
	Debug  bool   `mapstructure:"debug"`
	Server Server `mapstructure:"server"`
}
type Server struct {
	Name              string `mapstructure:"name"`
	Module            string
	DisableTelemetry  bool   `mapstructure:"disable_telemetry"`
	DisableDeployment bool   `mapstructure:"disable_deployment"`
	MetricsUrl        string `mapstructure:"metrics_url"`
	EnableNats        bool   `mapstructure:"enable_nats"`
	NatsSubject       string `mapstructure:"nats_subject"`
	NatsServers       string `mapstructure:"nats_servers"`
	EnableGraphql     bool   `mapstructure:"enable_graphql"`
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cwgo.yaml)")
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
		viper.SetConfigName(".cwgo")
	}

	viper.SetEnvPrefix("cwgo")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err == nil {
		logging.Debugf("using config %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		cobra.CheckErr(err)
	}

}

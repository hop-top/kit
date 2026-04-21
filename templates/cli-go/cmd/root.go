package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"{{module_prefix}}/{{app_name}}/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "{{app_name}}",
	Short:   "{{description}}",
	Version: version.Version(),
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP(
		"format", "f", "text",
		"Output format (text, json, yaml)",
	)
	rootCmd.PersistentFlags().BoolP(
		"verbose", "v", false,
		"Verbose output",
	)
}

func initConfig() {
	viper.SetEnvPrefix("{{app_name_upper}}")
	viper.AutomaticEnv()
	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(
			fmt.Sprintf("%s/.config/{{app_name}}", home),
		)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	_ = viper.ReadInConfig()
}

package main

import (
	"github.com/spf13/cobra"
	"main/pkg"
	configPkg "main/pkg/config"
	"main/pkg/fs"
	"main/pkg/logger"
)

var (
	version = "unknown"
)

func ExecuteMain(configPath string) {
	filesystem := &fs.OsFS{}
	app := pkg.NewApp(configPath, filesystem, version)
	app.Start()
}

func ExecuteValidateConfig(configPath string) {
	filesystem := &fs.OsFS{}

	config, err := configPkg.GetConfig(configPath, filesystem)
	if err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not load config!")
	}

	if err := config.Validate(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Config is invalid!")
	}

	logger.GetDefaultLogger().Info().Msg("Provided config is valid.")
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:     "missed-blocks-checker --config [config path]",
		Long:    "Monitors validators' missed blocks on Cosmos chains.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteMain(ConfigPath)
		},
	}

	validateConfigCmd := &cobra.Command{
		Use:     "validate-config --config [config path]",
		Long:    "Validate config.",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			ExecuteValidateConfig(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = rootCmd.MarkPersistentFlagRequired("config")

	validateConfigCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	_ = validateConfigCmd.MarkPersistentFlagRequired("config")

	rootCmd.AddCommand(validateConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetDefaultLogger().Panic().Err(err).Msg("Could not start application")
	}
}

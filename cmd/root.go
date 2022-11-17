package cmd

import (
	"fmt"
	"kdiff/internal/config"
	"kdiff/internal/helpers"
	"kdiff/internal/view"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName      = "kdiff"
	shortAppDesc = "A terminal UI tool for comparing resources across multiple kubernetes clusters."
	longAppDesc  = "kdiff is a terminal UI tool for comparing resources across multiple kubernetes clusters."
)

var (
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
		Run:   run,
	}
	version, commit = "dev", "dev"
	appConfig       config.ConfigFlags
)

func init() {
	// Set logging config
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)

	// Initialize config, commands and flags
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(versionCmd())
	initConfigFlags()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(config.DefaultConfigFile)

	// Bind cobra PFlags to viper.
	viper.BindPFlags(rootCmd.Flags())
	// Read in environment variables that match.
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initConfigFlags() {
	rootCmd.Flags().StringP(
		"logFile", "L",
		config.DefaultLogFile,
		"Specify the log file",
	)
	rootCmd.Flags().StringP(
		"logLevel", "l",
		config.DefaultLogLevel,
		"Specify a log level (info, warn, debug, trace, error)",
	)
	rootCmd.Flags().IntP(
		"refresh", "r",
		config.DefaultRefreshRate,
		"Specify the refresh rate (in seconds)",
	)
	rootCmd.Flags().StringP(
		"kubeconfig", "f",
		config.DefaultKubeconfig,
		"Path to the kubeconfig file",
	)
}

func run(cmd *cobra.Command, args []string) {
	// Get final values from Viper (order precedence: config file --> env var --> cmd line arg)
	appConfig = config.ConfigFlags{
		LogFile:     viper.GetString("logFile"),
		LogLevel:    viper.GetString("logLevel"),
		RefreshRate: viper.GetInt("refreshRate"),
		KubeConfig:  viper.GetString("kubeconfig"),
	}

	// Open log file for writing/appending
	config.EnsureDir(appConfig.LogFile, config.DefaultDirMod)
	fileOptions := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	logFile, err := os.OpenFile(appConfig.LogFile, fileOptions, config.DefaultFileMod)
	helpers.HandleError(err)
	defer logFile.Close()

	view.App(appConfig)
}

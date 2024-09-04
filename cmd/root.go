// Copyright © 2024 Triad National Security, LLC. All rights reserved.
//
// This program was produced under U.S. Government contract 89233218CNA000001
// for Los Alamos National Laboratory (LANL), which is operated by Triad
// National Security, LLC for the U.S. Department of Energy/National Nuclear
// Security Administration. All rights in the program are reserved by Triad
// National Security, LLC, and the U.S. Department of Energy/National Nuclear
// Security Administration. The Government is granted for itself and others
// acting on its behalf a nonexclusive, paid-up, irrevocable worldwide license
// in this material to reproduce, prepare derivative works, distribute copies to
// the public, perform publicly and display publicly, and to permit others to do
// so.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/config"
	"github.com/synackd/ochami/internal/log"
)

const (
	progName         = "ochami"
	defaultLogFormat = "json"
	defaultLogLevel  = "warning"
)

var (
	configFile   string
	configFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     progName,
	Short:   "Command line interface for interacting with OpenCHAMI services",
	Long:    "",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to execute root command")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(
		InitConfig,
		InitLogging,
	)
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "path to configuration file to use")
	rootCmd.PersistentFlags().StringVarP(&configFormat, "config-format", "", "", "format of configuration file; if none passed, tries to infer from file extension")
	rootCmd.PersistentFlags().String("log-format", defaultLogFormat, "log format (json,rfc3339,basic)")
	rootCmd.PersistentFlags().StringP("log-level", "l", defaultLogLevel, "set verbosity of logs (info,warning,debug)")

	checkBindError(viper.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("log-format")))
	checkBindError(viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")))
}

func checkBindError(e error) {
	if e != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to bind key to flag: %v\n", progName, e)
	}
}

func InitLogging() {
	// Set log level verbosity based on config file (log.level) or how many --log-level.
	// The command line option overrides the config file option.
	logCfg := viper.Sub("log")
	if logCfg == nil {
		fmt.Fprintf(os.Stderr, "%s: failed to read logging config", progName)
		os.Exit(1)
	}
	
	if err := log.Init(logCfg.GetString("level"), logCfg.GetString("format")); err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to initialize logger: %v\n", progName, err)
		os.Exit(1)
	}
}

func InitConfig() {
	// Set defaults for any keys not set by env var, config file, or flag
	config.SetDefaults()

	// Read any environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ochami")

	// Read configuration file if passed
	if configFile != "" {
		err := config.LoadConfig(configFile, configFormat)
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Fprintf(os.Stderr, "%s: configuration file %s not found: %v\n", progName, configFile, err)
			} else {
				fmt.Fprintf(os.Stderr, "%s: failed to load configuration file %s: %v\n", progName, configFile, err)
			}
			os.Exit(1)
		}
	}
}

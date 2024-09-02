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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/config"
	"github.com/synackd/ochami/internal/log"
)

const (
	progName = "ochami"
	defaultLogFormat = "json"
)

var (
	configFile   string
	configFormat string
	logFormat    string
	logLevel     int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   progName,
	Short: "Command line interface for interacting with OpenCHAMI services",
	Long:  "",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print help")
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to configuration file to use")
	rootCmd.PersistentFlags().StringVarP(&configFormat, "config-format", "", "", "Format of configuration file; if none passed, tries to infer from file extension")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", fmt.Sprintf("Log format (json,rfc3339,basic) (default: %s)", defaultLogFormat))
	rootCmd.PersistentFlags().CountVarP(&logLevel, "log-level", "l", "Set verbosity of logs; each additional -l increases the logging verbosity")

	checkBindError(viper.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("log-format")))
	checkBindError(viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")))
}

func checkBindError(e error) {
	if e != nil {
		//log.Logger.Error().Err(e).Msg("failed to bind key to flag")
		fmt.Fprintf(os.Stderr, "%s: failed to bind key to flag: %v\n", progName, e)
	}
}

func InitLogging() {
	// Set log level verbosity based on config file (log.level) or how many -l flags were passed.
	// The command line option overrides the config file option.
	var loggerLevel log.LogLevel
	if logLevel == 0 {
		if viper.IsSet("log.level") {
			ll := viper.GetString("log.level")
			switch ll {
			case "warning":
				loggerLevel = log.LogLevelWarning
			case "info":
				loggerLevel = log.LogLevelInfo
			case "debug":
				loggerLevel = log.LogLevelDebug
			default:
				fmt.Fprintf(os.Stderr, "%s: unknown log level %q\n", progName, ll)
				os.Exit(1)
			}
		} else {
			loggerLevel = log.LogLevelWarning
		}
	} else if logLevel == 1 {
		loggerLevel = log.LogLevelInfo
	} else if logLevel > 1 {
		loggerLevel = log.LogLevelDebug
	}

	// Set logging format based on config file (log.format) or --log-format.
	// The command line option overrides the config file option.
	var loggerFormat log.LogFormat
	if logFormat == "" {
		if viper.IsSet("log.format") {
			logFormat = viper.GetString("log.format")
		} else {
			logFormat = defaultLogFormat
		}
	}
	switch logFormat {
	case "rfc3339":
		loggerFormat = log.LogFormatRFC3339
	case "json":
		loggerFormat = log.LogFormatJSON
	case "basic":
		loggerFormat = log.LogFormatBasic
	default:
		fmt.Fprintf(os.Stderr, "%s: unknown log format %q\n", progName, logFormat)
		os.Exit(1)
	}

	if err := log.Init(loggerLevel, loggerFormat); err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to initialize logger: %v\n", progName, err)
		os.Exit(1)
	}
}

func InitConfig() {
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

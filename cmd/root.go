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
	"github.com/synackd/ochami/internal/log"
)

const (
	progName = "ochami"
)

var (
	configFile string
	logFormat  string
	logLevel   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   progName,
	Short: "Command line interface for interacting with OpenCHAMI services",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print help")
			}
			os.Exit(0)
		}

		// Set log level verbosity based on how many -l flags were passed.
		var loggerLevel log.LogLevel
		if logLevel == 0 {
			loggerLevel = log.LogLevelWarning
		} else if logLevel == 1 {
			loggerLevel = log.LogLevelInfo
		} else if logLevel > 1 {
			loggerLevel = log.LogLevelDebug
		}

		// Set logging format based on --log-level.
		var loggerFormat log.LogFormat
		switch logFormat {
		case "rfc3339":
			loggerFormat = log.LogFormatRFC3339
		case "json":
			loggerFormat = log.LogFormatJSON
		case "basic":
			loggerFormat = log.LogFormatBasic
		default:
			fmt.Fprintf(os.Stderr, "%s: unknown log format %q", progName, logFormat)
			os.Exit(1)
		}

		if err := log.Init(loggerLevel, loggerFormat); err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to initialize logger: %v\n", progName, err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to execute root command")
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "c", "Path to configuration file to use")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "json", "Log format (json,rfc3339,basic)")
	rootCmd.PersistentFlags().CountVarP(&logLevel, "log-level", "l", "Set verbosity of logs; each additional -l increases the logging verbosity")
}

func InitConfig() {
	if configFile != "" {
	} else {
		log.Logger.Debug().Msg("No configuration file passed on command line")
	}
}

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
	verbosity  int
	configFile string
	logFormat  string
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

		// Set log level verbosity based on how many -v flags were passed.
		var logLevel log.LogLevel
		if verbosity == 0 {
			logLevel = log.LogLevelWarning
		} else if verbosity == 1 {
			logLevel = log.LogLevelInfo
		} else if verbosity > 1 {
			logLevel = log.LogLevelDebug

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
		log.Logger.Warn().Msg("WARN")
		log.Logger.Info().Msg("INFO")
		log.Logger.Debug().Msg("DEBUG")
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
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Set verbosity of logs; each additional -v increases the verbosity")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "json", "Log format (json,rfc3339,basic)")
}

func InitConfig() {
	if configFile != "" {
	} else {
		log.Logger.Debug().Msg("No configuration file passed on command line")
	}
}

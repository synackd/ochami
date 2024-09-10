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
	"github.com/synackd/ochami/internal/version"
)

const (
	progName         = "ochami"
	defaultLogFormat = "json"
	defaultLogLevel  = "warning"
)

var (
	configFile   string
	configFormat string
	logLevel     string
	logFormat    string

	// These are only used by 'bss' and 'smd' subcommands.
	baseURI      string
	cacertPath   string
	token        string
	insecure     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     progName,
	Short:   "Command line interface for interacting with OpenCHAMI services",
	Long:    "",
	Version: version.Version,
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
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", defaultLogFormat, "log format (json,rfc3339,basic)")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", defaultLogLevel, "set verbosity of logs (info,warning,debug)")

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

	// Viper's BindPFlag does not currently work with binding to subkeys.
	// (See: https://github.com/spf13/viper/issues/368)
	// Therefore, we must manually check if the flag was set. If not, check if
	// config file option was set. If not, use default value.
	//
	// These if statements should be removed when the referenced issue is resolved.
	if !rootCmd.PersistentFlags().Lookup("log-format").Changed {
		if lf := logCfg.GetString("format"); lf != "" {
			logFormat = lf
		}
	}
	if !rootCmd.PersistentFlags().Lookup("log-level").Changed {
		if ll := logCfg.GetString("level"); ll != "" {
			logLevel = ll
		}
	}

	if err := log.Init(logLevel, logFormat); err != nil {
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

func getBaseURI(cmd *cobra.Command) (string, error) {
	// Precedence of getting base URI for requests:
	//
	// 1. If --cluster is set, search config file for matching name and read
	//    details from there.
	// 2. If flags corresponding to cluster info (e.g. --base-uri) are set,
	//    read details from them.
	// 3. If "default-cluster" is set in config file (config file must be
	//    specified), use cluster identified by that name as source of info.
	// 4. Data sources exhausted, err.
	var (
		clusterList  []map[string]any
		clusterToUse *map[string]any
		clusterName  string
	)
	if cmd.Flag("cluster").Changed {
		if configFile == "" {
			return "", fmt.Errorf("--cluster specified without --config")
		}
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			return "", fmt.Errorf("failed to unmarshal cluster list: %v", err)
		}
		clusterName = cmd.Flag("cluster").Value.String()
		for _, c := range clusterList {
			if c["name"] == clusterName {
				clusterToUse = &c
				break;
			}
		}
		if clusterToUse == nil {
			return "", fmt.Errorf("cluster %q not found in %s", clusterName, configFile)
		}
		clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
		if clusterToUseData["base-uri"] == nil {
			return "", fmt.Errorf("base-uri not set for cluster %q specified with --cluster", clusterName)
		}
		return clusterToUseData["base-uri"].(string), nil
	} else if cmd.Flag("base-uri").Changed {
		return baseURI, nil
	} else if configFile != "" && viper.IsSet("default-cluster") {
		clusterName = viper.GetString("default-cluster")
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			return "", fmt.Errorf("failed to unmarshal cluster list: %v", err)
		}
		for _, c := range clusterList {
			if c["name"] == clusterName {
				clusterToUse = &c
				break;
			}
		}
		if clusterToUse == nil {
			return "", fmt.Errorf("default cluster %q not found in %s", clusterName, configFile)
		}
		clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
		if clusterToUseData["base-uri"] == nil {
			return "", fmt.Errorf("base-uri not set for default cluster %q", clusterName)
		}
		return clusterToUseData["base-uri"].(string), nil
	}

	return "", fmt.Errorf("no base-uri set bia --base-uri, --cluster, or config file")
}

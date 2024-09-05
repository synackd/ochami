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
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/log"
)

// The 'show' subcommand of the 'config' command prints out the configuration
// values that the CLI sees, whether that be from a flag
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "View configuration options the CLI sees from a config file",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err          error
			cfgDataBytes []byte
		)
		cfgDataMap := viper.AllSettings()
		format := cmd.Flag("format").Value.String()
		switch format {
		case "yaml":
			cfgDataBytes, err = yaml.Marshal(&cfgDataMap)
		case "json":
			cfgDataBytes, err = json.MarshalIndent(&cfgDataMap, "", "\t")
		default:
			log.Logger.Error().Msgf("unknown log output format: %s", format)
			os.Exit(1)
		}
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal configuration data")
			os.Exit(1)
		}
		fmt.Println(string(cfgDataBytes))
	},
}

func init() {
	configShowCmd.Flags().StringP("format", "f", "yaml", "format of config output (yaml,json)")
	configCmd.AddCommand(configShowCmd)
}

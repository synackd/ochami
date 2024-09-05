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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/config"
	"github.com/synackd/ochami/internal/log"
)

// delCmd represents the del command
var configClusterDelCmd = &cobra.Command{
	Use:   "del CLUSTER_NAME",
	Short: "Delete a cluster from the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// Check that cluster name is only arg
		if len(args) == 0 {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		} else if len(args) > 1 {
			log.Logger.Error().Msgf("expected 1 argument (cluster name) but got %d: %v", len(args), args)
			os.Exit(1)
		}

		// We must have a config file in order to write cluster info
		if configFile == "" {
			log.Logger.Error().Msg("no config file path specified")
			os.Exit(1)
		}

		// Fetch existing cluster list config
		var clusterList []map[string]any // List of clusters in config
		clusterName := args[0]
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal cluster list")
		}
		for idx, cluster := range clusterList {
			if cluster["name"] == clusterName {
				newClusterList := config.RemoveFromSlice(clusterList, idx)

				// Apply config to Viper and write out the config file
				// WARNING: This will rewrite the whole config file so modifications like
				// comments will get erased.
				viper.Set("clusters", newClusterList)
				if err := viper.WriteConfig(); err != nil {
					log.Logger.Error().Err(err).Msgf("failed to write to config file: %s", configFile)
					os.Exit(1)
				}
				log.Logger.Info().Msgf("cluster %s removed from config file %s", clusterName, configFile)
				os.Exit(0)
			}
		}

		// If we have reached here, the cluster was not found
		log.Logger.Error().Msgf("cluster %s not found in config file %s", clusterName, configFile)
		os.Exit(1)
	},
}

func init() {
	configClusterCmd.AddCommand(configClusterDelCmd)
}

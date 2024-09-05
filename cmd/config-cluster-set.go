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
	"github.com/synackd/ochami/internal/log"
)

// setClusterCmd represents the setCluster command
var configClusterSetCmd = &cobra.Command{
	Use:   "set CLUSTER_NAME",
	Short: "Add or set parameters for a cluster",
	Long: `Use set-cluster to add cluster with its configuration or set the configuration
for an existing cluster. For example:

	ochami config set-cluster foobar.openchami.cluster --base-url https://foobar.openchami.cluster

Creates the following entry in the 'clusters' list:

	- name: foobar
	  cluster:
	    base-url: https://foobar.openchami.cluster

This same command can be use to modify existing cluster information. Running the same command above
with a different base URL will change the base URL for the 'foobar' cluster.`,
	Example: `  ochami config set-cluster foobar.openchami.cluster --base-url https://foobar.openchami.cluster`,
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

		var (
			clusterList []map[string]any // List of clusters in config
			modCluster  *map[string]any  // Pointer to existing cluster if not adding new
		)

		// Fetch existing cluster list config
		clusterName := args[0]
		clusterUrl := cmd.Flag("base-url").Value.String()
		if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal cluster list")
		}
		// If cluster name already exists, we are modifying it instead of creating a new one
		for _, cluster := range clusterList {
			if cluster["name"] == clusterName {
				modCluster = &cluster
				break
			}
		}

		if modCluster == nil {
			// Cluster does not exist, create a new entry for it in the config
			newCluster := make(map[string]any)
			newCluster["name"] = clusterName
			newClusterData := make(map[string]any)
			if clusterUrl != "" {
				newClusterData["base-url"] = clusterUrl
				log.Logger.Debug().Msgf("using base-url %s", clusterUrl)
			}
			newCluster["cluster"] = newClusterData
			clusterList = append(clusterList, newCluster)
			log.Logger.Info().Msgf("added new cluster: %s", clusterName)
		} else {
			// Cluster exists, modify it
			if clusterUrl != "" {
				modClusterData := (*modCluster)["cluster"].(map[string]any)
				modClusterData["base-url"] = clusterUrl
				(*modCluster)["cluster"] = modClusterData
				log.Logger.Debug().Msgf("updating base-url for cluster %s: %s", clusterName, clusterUrl)
			}
			log.Logger.Info().Msgf("modified config for existing cluster: %s", clusterName)
		}

		// Apply config to Viper and write out the config file
		// WARNING: This will rewrite the whole config file so modifications like
		// comments will get erased.
		viper.Set("clusters", clusterList)
		if err := viper.WriteConfig(); err != nil {
			log.Logger.Error().Err(err).Msgf("failed to write to config file: %s", configFile)
			os.Exit(1)
		}
		log.Logger.Info().Msgf("wrote config to %s", configFile)
	},
}

func init() {
	configClusterSetCmd.Flags().StringP("base-url", "u", "", "base URL of cluster")
	configClusterCmd.AddCommand(configClusterSetCmd)
}

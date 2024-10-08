// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/log"
)

// configClusterSetCmd represents the config-cluster-set command
var configClusterSetCmd = &cobra.Command{
	Use:   "set CLUSTER_NAME",
	Short: "Add or set parameters for a cluster",
	Long: `Use set-cluster to add cluster with its configuration or set the configuration
for an existing cluster. For example:

	ochami config cluster set foobar --base-uri https://foobar.openchami.cluster

Creates the following entry in the 'clusters' list:

	- name: foobar
	  cluster:
	    base-uri: https://foobar.openchami.cluster

If this is the first cluster created, the following is also set:

	default-cluster: foobar

default-cluster is used to determine which cluster in the list should be used for subcommands.

This same command can be use to modify existing cluster information. Running the same command above
with a different base URL will change the base URL for the 'foobar' cluster.`,
	Example: `  ochami config set-cluster foobar.openchami.cluster --base-uri https://foobar.openchami.cluster`,
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
		clusterUrl := cmd.Flag("base-uri").Value.String()
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
				newClusterData["base-uri"] = clusterUrl
				log.Logger.Debug().Msgf("using base-uri %s", clusterUrl)
			}
			newCluster["cluster"] = newClusterData

			// If this is the first cluster to be added, set it as the default
			if len(clusterList) == 0 {
				viper.Set("default-cluster", clusterName)
				log.Logger.Info().Msgf("first and new cluster %s set as default-cluster", clusterName)
			}

			// Add new cluster to list
			clusterList = append(clusterList, newCluster)
			log.Logger.Info().Msgf("added new cluster: %s", clusterName)

		} else {
			// Cluster exists, modify it
			if clusterUrl != "" {
				modClusterData := (*modCluster)["cluster"].(map[string]any)
				modClusterData["base-uri"] = clusterUrl
				(*modCluster)["cluster"] = modClusterData
				log.Logger.Debug().Msgf("updating base-uri for cluster %s: %s", clusterName, clusterUrl)
			}
			log.Logger.Info().Msgf("modified config for existing cluster: %s", clusterName)
		}

		// If --default was passed, make this cluster the default one
		if cmd.Flag("default").Changed {
			viper.Set("default-cluster", clusterName)
			log.Logger.Info().Msgf("cluster %s set as default-cluster due to --default being passed", clusterName)
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
	configClusterSetCmd.Flags().StringP("base-uri", "u", "", "base URL of cluster")
	configClusterSetCmd.Flags().BoolP("default", "d", false, "set cluster as the default")
	configClusterCmd.AddCommand(configClusterSetCmd)
}

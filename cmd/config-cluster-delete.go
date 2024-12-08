// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/OpenCHAMI/ochami/internal/config"
	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configClusterDeleteCmd represents the config-cluster-delete command
var configClusterDeleteCmd = &cobra.Command{
	Use:   "delete <cluster_name>",
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

				// Apply config to Viper
				viper.Set("clusters", newClusterList)

				// If cluster was default, remove default-cluster
				if viper.IsSet("default-cluster") {
					cn := viper.GetString("default-cluster")
					if cn == clusterName {
						viper.Set("default-cluster", "")
						log.Logger.Info().Msgf("cluster %s removed as default-cluster from config file %s", clusterName, configFile)
					}
				}

				// Write out config file
				// WARNING: This will rewrite the whole config file so modifications like
				// comments will get erased.
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
	configClusterCmd.AddCommand(configClusterDeleteCmd)
}

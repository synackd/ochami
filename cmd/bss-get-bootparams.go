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
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// bssGetBootparamsCmd represents the bootparams command
var bssGetBootparamsCmd = &cobra.Command{
	Use:   "bootparams",
	Short: "Get boot parameters for one or all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			clusterList  []map[string]any
			clusterToUse *map[string]any
			clusterName  string
			bssBaseURI   string
		)
		// Precedence of getting cluster information for requests:
		//
		// 1. If --cluster is set, search config file for matching name and read
		//    details from there.
		// 2. If flags corresponding to cluster info (e.g. --base-uri) are set,
		//    read details from them.
		// 3. If "default-cluster" is set in config file (config file must be
		//    specified), use cluster identified by that name as source of info.
		// 4. Data sources exhausted, err.
		if cmd.Flag("cluster").Changed {
			if configFile == "" {
				log.Logger.Error().Msg("config file required to use --cluster")
				os.Exit(1)
			}
			if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
				log.Logger.Error().Err(err).Msg("failed to unmarshal cluster list")
			}
			clusterName = cmd.Flag("cluster").Value.String()
			for _, c := range clusterList {
				if c["name"] == clusterName {
					clusterToUse = &c
					break;
				}
			}
			if clusterToUse == nil {
				log.Logger.Error().Msgf("cluster %q not found in %s", clusterName, configFile)
				os.Exit(1)
			}
			clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
			if clusterToUseData["base-uri"] == nil {
				log.Logger.Error().Msgf("base-uri not set for cluster %s", clusterName)
			}
			bssBaseURI = clusterToUseData["base-uri"].(string)
		} else if cmd.Flag("base-uri").Changed {
			bssBaseURI = baseURI
		} else if configFile != "" && viper.IsSet("default-cluster") {
			clusterName = viper.GetString("default-cluster")
			if err := viper.UnmarshalKey("clusters", &clusterList); err != nil {
				log.Logger.Error().Err(err).Msg("failed to unmarshal cluster list")
			}
			for _, c := range clusterList {
				if c["name"] == clusterName {
					clusterToUse = &c
					break;
				}
			}
			if clusterToUse == nil {
				log.Logger.Error().Msgf("default cluster %q not found in %s", clusterName, configFile)
				os.Exit(1)
			}
			clusterToUseData := (*clusterToUse)["cluster"].(map[string]any)
			if clusterToUseData["base-uri"] == nil {
				log.Logger.Error().Msgf("base-uri not set for default cluster %s", clusterName)
			}
			bssBaseURI = clusterToUseData["base-uri"].(string)
		} else {
			log.Logger.Error().Msg("no base-uri set via --base-uri or config file")
			if err := cmd.Usage(); err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
			}
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		// TODO: Check token validity/expiration
		if token == "" {
			log.Logger.Error().Msg("no token set")
			if err := cmd.Usage(); err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
			}
			os.Exit(1)
		}

		// Create client to make request to BSS
		bssClient, err := client.NewBSSClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		if cacertPath != "" {
			log.Logger.Debug().Msgf("Attempting to use CA certificate at %s", cacertPath)
			err = bssClient.UseCACert(cacertPath)
			if err != nil {
				log.Logger.Error().Err(err).Msgf("failed to load CA certificate %s: %v", cacertPath)
				os.Exit(1)
			}
		}

		// If no args specified, get all boot parameters
		if len(args) == 0 {
			data, err := bssClient.GetData("/bootparameters", token, nil)
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("BSS boot parameter request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to request boot parameters from BSS")
				}
				os.Exit(1)
			}

			fmt.Println(data)
		}
	},
}

func init() {
	bssGetCmd.AddCommand(bssGetBootparamsCmd)
}

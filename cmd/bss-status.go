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
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// bssStatusCmd represents the bss-status command
var bssStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get status of BSS service",
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		bssBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
			os.Exit(1)
		}

		// Create client to make request to BSS
		bssClient, err := client.NewBSSClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(bssClient.OchamiClient)

		// Determine which component to get status for and send request
		var httpEnv client.HTTPEnvelope
		if cmd.Flag("all").Changed {
			httpEnv, err = bssClient.GetStatus("all")
		} else if cmd.Flag("storage").Changed {
			httpEnv, err = bssClient.GetStatus("storage")
		} else if cmd.Flag("smd").Changed {
			httpEnv, err = bssClient.GetStatus("smd")
		} else if cmd.Flag("version").Changed {
			httpEnv, err = bssClient.GetStatus("version")
		} else {
			httpEnv, err = bssClient.GetStatus("")
		}
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS status request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to get BSS status")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	bssStatusCmd.Flags().Bool("all", false, "print all status data from BSS")
	bssStatusCmd.Flags().Bool("storage", false, "print status of storage backend from BSS")
	bssStatusCmd.Flags().Bool("smd", false, "print status of BSS connection to SMD")
	bssStatusCmd.Flags().Bool("version", false, "print version of BSS")

	bssStatusCmd.MarkFlagsMutuallyExclusive("all", "storage", "smd", "version")

	bssCmd.AddCommand(bssStatusCmd)
}

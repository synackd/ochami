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
	"github.com/synackd/ochami/internal/log"
)

// bssCmd represents the bss command
var bssCmd = &cobra.Command{
	Use:   "bss",
	Short: "Communicate with the Boot Script Service (BSS)",
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

func init() {
	bssCmd.PersistentFlags().String("cluster", "", "name of cluster whose config to use for this command")
	bssCmd.PersistentFlags().StringVarP(&baseURI, "base-uri", "u", "", "base URI for OpenCHAMI services")
	bssCmd.PersistentFlags().StringVar(&cacertPath, "cacert", "", "path to root CA certificate in PEM format")
	bssCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "access token to present for authentication")
	bssCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "do not verify TLS certificates")

	// Either use cluster from config file or specify details on CLI
	bssCmd.MarkFlagsMutuallyExclusive("cluster", "base-uri")

	if t, set := os.LookupEnv("OCHAMI_ACCESS_TOKEN"); set {
		if !bssCmd.PersistentFlags().Lookup("token").Changed {
			token = t
		}
	}
	rootCmd.AddCommand(bssCmd)
}

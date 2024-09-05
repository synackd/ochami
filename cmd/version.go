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

	"github.com/spf13/cobra"
)

var (
	version = "v0.0.0"
	commit  = "000000"
	date    = "0000-00-00:00:00:00"
	output  string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version to stdout and exit",
	Example: `  ochami version
  ochami version --all
  ochami version -c`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("all").Value.String() == "true" {
			output = fmt.Sprintf("%s %s @ %s", version, commit, date)
		} else if cmd.Flag("commit").Value.String() == "true" {
			output = fmt.Sprintf("%s @ %s", commit, date)
		} else {
			output = version
		}
		fmt.Println(output)
	},
}

func init() {
	versionCmd.Flags().Bool("commit", false, "print just git commit and build date")
	versionCmd.Flags().BoolP("all", "a", false, "print version, git commit, and build date")
	versionCmd.MarkFlagsMutuallyExclusive("all", "commit")
	rootCmd.AddCommand(versionCmd)
}

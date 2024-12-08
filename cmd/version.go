// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"fmt"

	"github.com/OpenCHAMI/ochami/internal/version"
	"github.com/spf13/cobra"
)

var output string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Args:  cobra.NoArgs,
	Short: "Print version to stdout and exit",
	Example: `  ochami version
  ochami version --all
  ochami version -c`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("all").Value.String() == "true" {
			output = fmt.Sprintf("%s %s @ %s", version.Version, version.Commit, version.Date)
		} else if cmd.Flag("commit").Value.String() == "true" {
			output = fmt.Sprintf("%s @ %s", version.Commit, version.Date)
		} else {
			output = version.Version
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

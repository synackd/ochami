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
	Short: "Print detailed version to stdout and exit",
	Example: `  ochami version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", version.Version)
		fmt.Printf("Tag:        %s\n", version.Tag)
		fmt.Printf("Branch:     %s\n", version.Branch)
		fmt.Printf("Commit:     %s\n", version.Commit)
		fmt.Printf("Git State:  %s\n", version.GitState)
		fmt.Printf("Date:       %s\n", version.Date)
		fmt.Printf("Go:         %s\n", version.GoVersion)
		fmt.Printf("Build Host: %s\n", version.BuildHost)
		fmt.Printf("Build User: %s\n", version.BuildUser)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

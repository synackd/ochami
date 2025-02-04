//go:build magellan || all
// +build magellan all

package cmd

import (
	magellan_cmd "github.com/OpenCHAMI/magellan/cmd"
)

func init() {
	discoverCmd.AddCommand(magellan_cmd.ScanCmd)
	discoverCmd.AddCommand(magellan_cmd.CollectCmd)
	discoverCmd.AddCommand(magellan_cmd.ListCmd)
	discoverCmd.AddCommand(magellan_cmd.CrawlCmd)
}

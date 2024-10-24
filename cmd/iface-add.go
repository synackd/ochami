// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// ifaceAddCmd represents the iface-add command
var ifaceAddCmd = &cobra.Command{
	Use:   "add (-f <payload_file> | <comp_id> <mac_addr> (<net_name>,<ip_addr>)...)",
	Short: "Add new ethernet interface(s)",
	Long: `Add new ethernet interface(s). A component ID (usually an xname), MAC address, and
one or more pairs of network name and IP address (delimited by a comma) are required unless -f is
passed to read from a payload file. Specifying -f also is mutually exclusive with the other flags
of this command and its arguments.

This command sends a POST to SMD. An access token is required.`,
	Example: `  ochami iface add x3000c1s7b55n0 de:ca:fc:0f:fe:ee NMN,172.16.0.55
  ochami iface add -d "Node Management for n55" x3000c1s7b55n0 de:ca:fc:0f:fe:ee NMN,172.16.0.55
  ochami iface add x3000c1s7b55n0 de:ca:fc:0f:fe:ee external,10.1.0.55 internal,172.16.0.55
  ochami iface add -f payload.json
  ochami iface add -f payload.yaml --payload-format yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check that all required args are passed
		if len(args) == 0 && !cmd.Flag("payload").Changed {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		} else if len(args) < 3 {
			log.Logger.Error().Msgf("expected at least 3 arguments (comp_id, mac_addr, net_ip_paor) but got %d: %v", len(args), args)
			os.Exit(1)
		}

		// Without a base URI, we cannot do anything
		smdBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for SMD")
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		setTokenFromEnvVar(cmd)
		checkToken(cmd)

		// Create client to make request to SMD
		smdClient, err := client.NewSMDClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		var eis []client.EthernetInterface
		if cmd.Flag("payload").Changed {
			// Use payload file if passed
			dFile := cmd.Flag("payload").Value.String()
			dFormat := cmd.Flag("payload-format").Value.String()
			err := client.ReadPayload(dFile, dFormat, &eis)
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to read payload for request")
				os.Exit(1)
			}
		} else {
			// ...otherwise use CLI options/args
			var nets []client.EthernetIP
			for i := 2; i < len(args); i++ {
				tokens := strings.SplitN(args[i], ",", 2)
				if ip := net.ParseIP(tokens[1]); ip.To4() == nil {
					log.Logger.Error().Msgf("invalid IP address: %s", tokens[1])
					os.Exit(1)
				}
				net := client.EthernetIP{
					Network:   tokens[0],
					IPAddress: tokens[1],
				}
				nets = append(nets, net)
			}
			ei := client.EthernetInterface{
				ComponentID: args[0],
				Description: cmd.Flag("description").Value.String(),
				MACAddress:  args[1],
				IPAddresses: nets,
			}
			eis = append(eis, ei)
		}

		// Send off request
		_, errs, err := smdClient.PostEthernetInterfaces(eis, token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to add ethernet interface in SMD")
			os.Exit(1)
		}
		// Since smdClient.PostEthernetInterfaces does the addition iteratively, we need to deal with
		// each error that might have occurred.
		var errorsOccurred = false
		for _, err := range errs {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD ethernet interface request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to add ethernet interfaces to BSS")
			}
		}
		if errorsOccurred {
			log.Logger.Warn().Msg("SMD ethernet interface addition completed with errors")
			os.Exit(1)
		}
	},
}

func init() {
	ifaceAddCmd.Flags().StringP("description", "d", "Undescribed Ethernet Interface", "description of interface")
	ifaceAddCmd.Flags().StringP("payload", "f", "", "file containing the request payload; JSON format unless --payload-format specified")
	ifaceAddCmd.Flags().String("payload-format", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")

	ifaceAddCmd.MarkFlagsMutuallyExclusive("description", "payload")

	ifaceCmd.AddCommand(ifaceAddCmd)
}
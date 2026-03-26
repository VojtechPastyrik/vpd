package add

import (
	"strings"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var hostsFile string

var Cmd = &cobra.Command{
	Use:   "add <profile> <ip> <hostname> [aliases...]",
	Short: "Add a host entry to a profile",
	Long: `Add a host entry to a profile. Creates the profile if it doesn't exist.

Examples:
  vpd hosts add myproject 127.0.0.1 myapp.local
  vpd hosts add myproject 127.0.0.1 myapp.local api.local cdn.local`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]
		ip := args[1]
		hostname := args[2]
		var aliases []string
		if len(args) > 3 {
			aliases = args[3:]
		}

		hf, err := hostsutil.Parse(hostsFile)
		if err != nil {
			logger.Fatalf("failed to parse hosts file: %v", err)
		}

		entry := hostsutil.Entry{
			IP:       ip,
			Hostname: hostname,
			Aliases:  aliases,
		}
		hf.AddEntries(profileName, []hostsutil.Entry{entry})

		if err := hf.Write(hostsFile); err != nil {
			logger.Fatalf("failed to write hosts file: %v", err)
		}

		all := append([]string{hostname}, aliases...)
		logger.Successf("added %s → %s to profile %q", ip, strings.Join(all, ", "), profileName)
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	parent_cmd.Cmd.AddCommand(Cmd)
}

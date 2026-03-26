package remove

import (
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var (
	hostsFile string
	hostname  string
)

var Cmd = &cobra.Command{
	Use:   "remove <profile> [--hostname <hostname>]",
	Short: "Remove a profile or a single entry from a profile",
	Long: `Remove an entire profile or a specific hostname entry from a profile.

Examples:
  vpd hosts remove myproject                       # remove entire profile
  vpd hosts remove myproject --hostname myapp.local # remove single entry`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		hf, err := hostsutil.Parse(hostsFile)
		if err != nil {
			logger.Fatalf("failed to parse hosts file: %v", err)
		}

		if hostname != "" {
			if !hf.RemoveEntry(profileName, hostname) {
				logger.Fatalf("entry %q not found in profile %q", hostname, profileName)
			}
			if err := hf.Write(hostsFile); err != nil {
				logger.Fatalf("failed to write hosts file: %v", err)
			}
			logger.Successf("removed %q from profile %q", hostname, profileName)
			return
		}

		if !hf.RemoveProfile(profileName) {
			logger.Fatalf("profile %q not found", profileName)
		}
		if err := hf.Write(hostsFile); err != nil {
			logger.Fatalf("failed to write hosts file: %v", err)
		}
		logger.Successf("removed profile %q", profileName)
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	Cmd.Flags().StringVar(&hostname, "hostname", "", "Remove a specific hostname instead of the entire profile")
	parent_cmd.Cmd.AddCommand(Cmd)
}

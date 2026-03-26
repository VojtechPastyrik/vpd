package disable

import (
	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var hostsFile string

var Cmd = &cobra.Command{
	Use:   "disable <profile>",
	Short: "Disable a hosts profile (comments out entries)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hf, err := hostsutil.Parse(hostsFile)
		if err != nil {
			logger.Fatalf("failed to parse hosts file: %v", err)
		}

		if !hf.DisableProfile(args[0]) {
			logger.Fatalf("profile %q not found", args[0])
		}

		if err := hf.Write(hostsFile); err != nil {
			logger.Fatalf("failed to write hosts file: %v", err)
		}
		logger.Successf("disabled profile %q", args[0])
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	parent_cmd.Cmd.AddCommand(Cmd)
}

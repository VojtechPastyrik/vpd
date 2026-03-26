package restore

import (
	"os"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var hostsFile string

var Cmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore the hosts file from a backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			logger.Fatalf("failed to read backup file: %v", err)
		}

		if err := os.WriteFile(hostsFile, data, 0644); err != nil {
			logger.Fatalf("failed to write hosts file: %v", err)
		}
		logger.Successf("hosts file restored from %s", args[0])
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	parent_cmd.Cmd.AddCommand(Cmd)
}

package backup

import (
	"fmt"
	"os"
	"time"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var (
	hostsFile  string
	outputPath string
)

var Cmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the hosts file",
	Run: func(cmd *cobra.Command, args []string) {
		dest := outputPath
		if dest == "" {
			dest = fmt.Sprintf("%s.%s.bak", hostsFile, time.Now().Format("20060102-150405"))
		}

		data, err := os.ReadFile(hostsFile)
		if err != nil {
			logger.Fatalf("failed to read hosts file: %v", err)
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			logger.Fatalf("failed to write backup: %v", err)
		}
		logger.Successf("backup saved to %s", dest)
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	Cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path (default: <hosts>.<timestamp>.bak)")
	parent_cmd.Cmd.AddCommand(Cmd)
}

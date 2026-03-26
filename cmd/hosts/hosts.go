package hosts

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "hosts",
	Short: "Manage /etc/hosts file profiles",
}

func init() {
	root.RootCmd.AddCommand(Cmd)
}

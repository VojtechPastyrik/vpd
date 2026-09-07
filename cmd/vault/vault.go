package vault

import (
	"github.com/VojtechPastyrik/vpd/cmd/root"
	"github.com/spf13/cobra"
)

var FlagFlavor string

var Cmd = &cobra.Command{
	Use:     "vault",
	Short:   "vault / OpenBao Utils",
	Aliases: []string{"openbao", "bao"},
}

func init() {
	root.RootCmd.AddCommand(Cmd)
	Cmd.PersistentFlags().StringVar(
		&FlagFlavor,
		"flavor",
		"auto",
		"Server flavor: auto, vault or openbao (auto detects by pod labels)",
	)
}

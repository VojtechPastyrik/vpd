package list

import (
	"fmt"
	"os"
	"text/tabwriter"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/hosts"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	hostsutil "github.com/VojtechPastyrik/vpd/utils/hosts"
	"github.com/spf13/cobra"
)

var hostsFile string

var Cmd = &cobra.Command{
	Use:   "list [profile]",
	Short: "List profiles and their host entries",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hf, err := hostsutil.Parse(hostsFile)
		if err != nil {
			logger.Fatalf("failed to parse hosts file: %v", err)
		}

		if len(hf.Profiles) == 0 {
			fmt.Println("No profiles found.")
			return
		}

		if len(args) == 1 {
			p := hf.GetProfile(args[0])
			if p == nil {
				logger.Fatalf("profile %q not found", args[0])
			}
			printProfile(p)
			return
		}

		for i, p := range hf.Profiles {
			if i > 0 {
				fmt.Println()
			}
			printProfile(&p)
		}
	},
}

func init() {
	Cmd.Flags().StringVarP(&hostsFile, "file", "f", hostsutil.DefaultPath(), "Path to hosts file")
	parent_cmd.Cmd.AddCommand(Cmd)
}

func printProfile(p *hostsutil.Profile) {
	status := "on"
	if !p.Enabled {
		status = "off"
	}
	fmt.Printf("Profile: %s [%s]\n", p.Name, status)
	if len(p.Entries) == 0 {
		fmt.Println("  (no entries)")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, e := range p.Entries {
		fmt.Fprintf(w, "  %s\t%s\n", e.IP, e.Hostname)
	}
	w.Flush()
}

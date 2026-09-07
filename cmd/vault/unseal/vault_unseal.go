package unseal

import (
	"os"
	"os/exec"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/vault"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	vault_utils "github.com/VojtechPastyrik/vpd/utils/vault"
	"github.com/spf13/cobra"
)

var flagFile string
var flagOpSecret string
var flagNamespace string
var flagPod string

var Cmd = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal Vault or OpenBao pods using keys from a file or 1Password",
	Long: `This command unseals Vault or OpenBao pods using unseal keys from a JSON file or a 1Password secret reference.
Useful when pods have restarted and need to be unsealed without re-initializing.
The flavor is detected from pod labels unless --flavor is set.`,
	Example: `  vpd vault unseal --file /path/to/vault_keys.json --namespace vault
  vpd vault unseal --op-secret "op://vault/keys/vault_keys" --namespace vault
  vpd vault unseal --file /path/to/vault_keys.json --namespace vault --pod vault-0
  vpd openbao unseal --file /path/to/vault_keys.json --namespace openbao --flavor openbao`,
	Args:    cobra.NoArgs,
	Aliases: []string{"us"},
	Run: func(cmd *cobra.Command, args []string) {
		vaultUnseal()
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&flagFile,
		"file",
		"f",
		"",
		"Path to vault_keys.json file",
	)
	Cmd.Flags().StringVarP(
		&flagOpSecret,
		"op-secret",
		"o",
		"",
		"1Password secret reference (e.g. op://vault/keys/vault_keys)",
	)
	Cmd.Flags().StringVarP(
		&flagNamespace,
		"namespace",
		"n",
		"vault",
		"Namespace where Vault is running",
	)
	Cmd.Flags().StringVarP(
		&flagPod,
		"pod",
		"P",
		"",
		"Specific pod name to unseal (default: all vault pods)",
	)
}

func vaultUnseal() {
	if flagFile == "" && flagOpSecret == "" {
		logger.Fatal("one of --file or --op-secret must be provided")
	}
	if flagFile != "" && flagOpSecret != "" {
		logger.Fatal("only one of --file or --op-secret can be provided")
	}

	var jsonData string
	if flagFile != "" {
		jsonData = loadKeysFromFile(flagFile)
	} else {
		jsonData = loadKeysFromOp(flagOpSecret)
	}

	keys, threshold := vault_utils.ExtractVaultKeys(jsonData)

	flavor, podNames := vault_utils.ResolveFlavor(parent_cmd.FlagFlavor, flagNamespace)
	if flagPod != "" {
		podNames = []string{flagPod}
	} else if len(podNames) == 0 {
		logger.Fatalf("no %s pods found in namespace %s", flavor.Name, flagNamespace)
	}

	for _, podName := range podNames {
		logger.Infof("unsealing pod %s", podName)
		vault_utils.UnsealPod(podName, flagNamespace, flavor, keys, threshold)
	}

	logger.Infof("%s unsealing completed successfully", flavor.Name)
}

func loadKeysFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Fatalf("error reading keys file: %v", err)
	}
	return string(data)
}

func loadKeysFromOp(secretRef string) string {
	cmd := exec.Command("op", "read", secretRef)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error reading secret from 1Password: %v\nOutput: %s", err, string(output))
	}
	return string(output)
}

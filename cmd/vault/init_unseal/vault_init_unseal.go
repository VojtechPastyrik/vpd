package init_unseal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	parent_cmd "github.com/VojtechPastyrik/vpd/cmd/vault"
	"github.com/VojtechPastyrik/vpd/pkg/logger"
	vault_utils "github.com/VojtechPastyrik/vpd/utils/vault"
	"github.com/spf13/cobra"
)

var FlagPath string
var FlagNamespace string
var FlagKeyShares int
var FlagKeyThreshold int

var Cmd = &cobra.Command{
	Use:   "init-unseal",
	Short: "Initialize and unseal Vault or OpenBao",
	Long: `This command initializes Vault or OpenBao and unseals it using the generated keys.
It requires a running instance in the specified namespace and saves the keys to the specified path.
The flavor is detected from pod labels unless --flavor is set.`,
	Example: `  vpd vault init-unseal --path /path/to/save/keys --namespace vault
  vpd openbao init-unseal --path /path/to/save/keys --namespace openbao --flavor openbao`,
	Args:    cobra.NoArgs,
	Aliases: []string{"iu"},
	Run: func(cmd *cobra.Command, args []string) {
		vaultInitUnseal(FlagPath, FlagNamespace, FlagKeyShares, FlagKeyThreshold)
	},
}

func init() {
	parent_cmd.Cmd.AddCommand(Cmd)
	Cmd.Flags().StringVarP(
		&FlagPath,
		"path",
		"p",
		"",
		"Path to save vault keys",
	)
	Cmd.MarkFlagRequired("path")
	Cmd.Flags().StringVarP(
		&FlagNamespace,
		"namespace",
		"n",
		"vault",
		"Namespace where Vault is running",
	)
	Cmd.Flags().IntVarP(
		&FlagKeyShares,
		"key-shares",
		"s",
		5,
		"Number of key shares to generate during Vault initialization",
	)
	Cmd.Flags().IntVarP(
		&FlagKeyThreshold,
		"key-threshold",
		"t",
		3,
		"Number of key shares required to unseal Vault",
	)
}

func vaultInitUnseal(path, namespace string, keyShares, keyThreshold int) {
	if keyShares < 1 || keyThreshold < 1 {
		logger.Fatal("key shares and threshold must be greater than 0")
	}
	if keyThreshold > keyShares {
		logger.Fatal("key threshold cannot be greater than key shares")
	}

	flavor, podNames := vault_utils.ResolveFlavor(parent_cmd.FlagFlavor, namespace)
	if len(podNames) == 0 {
		logger.Fatalf("no %s pods found in namespace %s", flavor.Name, namespace)
	}
	vaultKeys := vaultInit(podNames[0], path, namespace, flavor, keyShares, keyThreshold)

	extractedValueKeys, threshold := vault_utils.ExtractVaultKeys(vaultKeys)

	for _, podName := range podNames {
		logger.Infof("unsealing pod %s", podName)
		vault_utils.UnsealPod(podName, namespace, flavor, extractedValueKeys, threshold)
	}

	logger.Infof("%s initialization and unsealing completed successfully. Keys saved to %s/vault_keys.json\n", flavor.Name, path)
}

func vaultInit(pod, path, namespace string, flavor vault_utils.Flavor, keyShares, keyThreshold int) string {
	logger.Infof("executing %s operator init on pod %s in namespace %s\n", flavor.Binary, pod, namespace)
	cmd := exec.Command("kubectl", "exec", pod, "-n", namespace, "--", flavor.Binary, "operator", "init", "-format=json", "-key-shares", fmt.Sprintf("%d", keyShares), "-key-threshold", fmt.Sprintf("%d", keyThreshold))
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error executing %s operator init: %s\n", flavor.Binary, string(output))
	}

	err = os.WriteFile(filepath.Join(path, "vault_keys.json"), output, 0644)
	if err != nil {
		logger.Infof("[WARNING] error writing output to file: %v", err)
		logger.Infof("output: %s\n", string(output))
	}

	return string(output)
}

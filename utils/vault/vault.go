package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/VojtechPastyrik/vpd/pkg/logger"
)

type Flavor struct {
	Name     string
	PodLabel string
	Binary   string
}

var (
	FlavorVault   = Flavor{Name: "vault", PodLabel: "app.kubernetes.io/name=vault,component=server", Binary: "vault"}
	FlavorOpenBao = Flavor{Name: "openbao", PodLabel: "app.kubernetes.io/name=openbao,component=server", Binary: "bao"}
)

var flavors = []Flavor{FlavorVault, FlavorOpenBao}

// ResolveFlavor picks the flavor named by name and lists its server pods.
// With "auto" it returns the first flavor that has pods in the namespace.
func ResolveFlavor(name, namespace string) (Flavor, []string) {
	if name == "auto" {
		for _, f := range flavors {
			if pods := GetPods(namespace, f); len(pods) > 0 {
				return f, pods
			}
		}
		logger.Fatalf("no Vault or OpenBao pods found in namespace %s", namespace)
	}
	for _, f := range flavors {
		if f.Name == name {
			return f, GetPods(namespace, f)
		}
	}
	logger.Fatalf("unknown flavor %q (expected auto, vault or openbao)", name)
	return Flavor{}, nil
}

func GetPods(namespace string, flavor Flavor) []string {
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace,
		"-l", flavor.PodLabel,
		"-o", "jsonpath={.items[*].metadata.name}")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error executing kubectl command: %v output %s", err, string(output))
	}

	pods := strings.Split(strings.TrimSpace(string(output)), " ")
	var podNames []string
	for _, pod := range pods {
		if pod != "" {
			podNames = append(podNames, pod)
		}
	}

	return podNames
}

func ExtractVaultKeys(data string) ([]string, int) {
	var response struct {
		UnsealKeysB64   []string `json:"unseal_keys_b64"`
		UnsealThreshold int      `json:"unseal_threshold"`
	}

	err := json.Unmarshal([]byte(data), &response)
	if err == nil {
		return response.UnsealKeysB64, response.UnsealThreshold
	}

	keys, parseErr := ParseVaultKeysText(data)
	if parseErr != nil {
		logger.Fatalf("error parsing vault keys (tried JSON and text format): %v", parseErr)
	}

	return keys, len(keys)
}

func ParseVaultKeysText(text string) ([]string, error) {
	re := regexp.MustCompile(`(?m)^Unseal Key \d+:\s*(.+)$`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no unseal keys found in text")
	}

	var keys []string
	for _, match := range matches {
		keys = append(keys, strings.TrimSpace(match[1]))
	}

	return keys, nil
}

func UnsealPod(podName, namespace string, flavor Flavor, vaultKeys []string, threshold int) {
	for i, key := range vaultKeys {
		if i >= threshold {
			break
		}
		unsealWithRetry(podName, namespace, flavor, key)
	}
	WaitForPodReady(podName, namespace)
}

func unsealWithRetry(podName, namespace string, flavor Flavor, key string) {
	const maxRetries = 6
	const retryInterval = 5 * time.Second

	for attempt := range maxRetries {
		cmd := exec.Command("kubectl", "exec", podName, "-n", namespace, "--", flavor.Binary, "operator", "unseal", key)
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err == nil {
			return
		}
		if strings.Contains(string(output), "not initialized") && attempt < maxRetries-1 {
			logger.Infof("pod %s not yet initialized, retrying in %s (%d/%d)", podName, retryInterval, attempt+1, maxRetries)
			time.Sleep(retryInterval)
			continue
		}
		logger.Fatalf("error unsealing pod %s with key %s: %v\nOutput: %s", podName, key, err, output)
	}
}

func WaitForPodReady(pod, namespace string) {
	cmd := exec.Command("kubectl", "wait", "pod", pod, "-n", namespace, "--for=condition=Ready", "--timeout=60s")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Fatalf("error waiting for pod to be ready: %v\nOutput: %s", err, output)
	}
}

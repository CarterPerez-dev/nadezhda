// ©AngelaMos | 2026
// ollama.go

package setup

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed ollama.compose.yml
var ComposeYAML []byte

const (
	composeFileName = "ollama.compose.yml"
	composePerm     = 0o644
	versionPath     = "/api/version"
	apiV1Suffix     = "/v1"
)

var ollamaCandidates = []string{
	"http://localhost:39847",
	"http://localhost:11434",
}

func DetectOllama(client *http.Client) string {
	for _, base := range ollamaCandidates {
		resp, err := client.Get(base + versionPath)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return base
		}
	}
	return ""
}

func WriteCompose(dir string) (string, error) {
	path := filepath.Join(dir, composeFileName)
	if err := os.WriteFile(path, ComposeYAML, composePerm); err != nil {
		return "", fmt.Errorf("setup: write compose: %w", err)
	}
	return path, nil
}

func HasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func OllamaReachable(baseURL string) bool {
	base := strings.TrimSuffix(baseURL, apiV1Suffix)
	client := &http.Client{Timeout: detectTimeout}
	resp, err := client.Get(base + versionPath)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

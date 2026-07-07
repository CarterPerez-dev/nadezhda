// ©AngelaMos | 2026
// ollama_test.go

package setup

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDetectOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == versionPath {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	old := ollamaCandidates
	defer func() { ollamaCandidates = old }()
	client := &http.Client{Timeout: 2 * time.Second}

	ollamaCandidates = []string{srv.URL}
	if got := DetectOllama(client); got != srv.URL {
		t.Errorf("DetectOllama = %q, want %q", got, srv.URL)
	}

	ollamaCandidates = []string{"http://127.0.0.1:1"}
	if got := DetectOllama(client); got != "" {
		t.Errorf("DetectOllama with no server = %q, want empty", got)
	}
}

func TestWriteComposeMatchesEmbed(t *testing.T) {
	yaml := string(ComposeYAML)
	if len(yaml) == 0 || !strings.Contains(yaml, "nadezhda-ollama") || !strings.Contains(yaml, "qwen2.5:7b") {
		t.Fatal("embedded compose is empty or missing expected content")
	}

	dir := t.TempDir()
	path, err := WriteCompose(dir)
	if err != nil {
		t.Fatalf("WriteCompose: %v", err)
	}
	if path != filepath.Join(dir, composeFileName) {
		t.Errorf("path = %q", path)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != yaml {
		t.Error("written compose differs from embedded content")
	}
}

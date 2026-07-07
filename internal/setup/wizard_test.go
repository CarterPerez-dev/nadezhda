// ©AngelaMos | 2026
// wizard_test.go

package setup

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWizardAPIProvider(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var out bytes.Buffer
	in := strings.NewReader("2\nsk-test-key\n\n")

	if err := Run(in, &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	path, _ := CredentialsPath()
	kv, _ := readCredentials(path)
	if kv[EnvProvider] != ProviderOpenAI || kv["OPENAI_API_KEY"] != "sk-test-key" {
		t.Errorf("saved = %v, want openai + key", kv)
	}
	if _, ok := kv["NVD_API_KEY"]; ok {
		t.Error("NVD key should be absent when skipped")
	}
	if !strings.Contains(out.String(), "OpenAI is ready") {
		t.Errorf("output missing confirmation:\n%s", out.String())
	}
	if strings.Contains(out.String(), "sk-test-key") {
		t.Error("wizard echoed the API key to its output")
	}
}

func TestWizardOllamaDetected(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	old := ollamaCandidates
	ollamaCandidates = []string{srv.URL}
	defer func() { ollamaCandidates = old }()

	var out bytes.Buffer
	if err := Run(strings.NewReader("4\n"), &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	path, _ := CredentialsPath()
	kv, _ := readCredentials(path)
	if kv[EnvProvider] != ProviderQwen || kv[EnvQwenURL] != srv.URL+"/v1" {
		t.Errorf("saved = %v, want qwen + %s/v1", kv, srv.URL)
	}
	if !strings.Contains(out.String(), "Found a running Ollama") {
		t.Errorf("output missing detect line:\n%s", out.String())
	}
}

func TestWizardInvalidChoice(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Run(strings.NewReader("9\n"), &bytes.Buffer{}); err == nil {
		t.Error("invalid choice should error")
	}
}

func TestReadLine(t *testing.T) {
	cases := map[string]string{
		"hello\n":    "hello",
		"trim  \r\n": "trim",
		"noeol":      "noeol",
		"\n":         "",
	}
	for in, want := range cases {
		if got := readLine(strings.NewReader(in)); got != want {
			t.Errorf("readLine(%q) = %q, want %q", in, got, want)
		}
	}
}

// ©AngelaMos | 2026
// credentials_test.go

package setup

import (
	"os"
	"testing"
)

func TestCredentialsRoundTripAndPerms(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := Save(map[string]string{"OPENAI_API_KEY": "sk-abc", EnvProvider: "openai"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Save(map[string]string{"NVD_API_KEY": "nvd-1"}); err != nil {
		t.Fatalf("Save merge: %v", err)
	}

	path, _ := CredentialsPath()
	kv, err := readCredentials(path)
	if err != nil {
		t.Fatalf("readCredentials: %v", err)
	}
	if kv["OPENAI_API_KEY"] != "sk-abc" || kv[EnvProvider] != "openai" || kv["NVD_API_KEY"] != "nvd-1" {
		t.Errorf("merge lost data: %v", kv)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("credentials perm = %o, want 600", perm)
	}
}

func TestLoadEnvWins(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(map[string]string{"OPENAI_API_KEY": "fromfile", "NADEZHDA_TEST_ONLY": "x"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENAI_API_KEY", "fromenv")
	defer os.Unsetenv("NADEZHDA_TEST_ONLY")

	if err := Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := os.Getenv("OPENAI_API_KEY"); got != "fromenv" {
		t.Errorf("OPENAI_API_KEY = %q, want fromenv (env must win over file)", got)
	}
	if got := os.Getenv("NADEZHDA_TEST_ONLY"); got != "x" {
		t.Errorf("NADEZHDA_TEST_ONLY = %q, want x (unset var loads from file)", got)
	}
}

func TestReadCredentialsSkipsCommentsAndBlanks(t *testing.T) {
	path := t.TempDir() + "/creds"
	content := "# header\n\nOPENAI_API_KEY=sk-1\n  # indented comment\nbadline_no_sep\nNVD_API_KEY = nvd-2 \n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	kv, err := readCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if kv["OPENAI_API_KEY"] != "sk-1" || kv["NVD_API_KEY"] != "nvd-2" || len(kv) != 2 {
		t.Errorf("parsed = %v, want exactly 2 clean entries", kv)
	}
}

func TestSaveCreatesDir0700AndFile0600(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(map[string]string{"OPENAI_API_KEY": "k"}); err != nil {
		t.Fatal(err)
	}
	dir, _ := ConfigDir()
	if di, _ := os.Stat(dir); di.Mode().Perm() != 0o700 {
		t.Errorf("config dir perms = %o, want 700", di.Mode().Perm())
	}
	path, _ := CredentialsPath()
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Errorf("credentials perms = %o, want 600", fi.Mode().Perm())
	}
}

func TestSaveTightensExistingLooseFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir, _ := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path, _ := CredentialsPath()
	if err := os.WriteFile(path, []byte("OPENAI_API_KEY=old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(map[string]string{"OPENAI_API_KEY": "new"}); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Errorf("Save left perms %o on a pre-existing 0644 file, want 600", fi.Mode().Perm())
	}
}

func TestLoadRespectsSetEmptyEnv(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := Save(map[string]string{"OPENAI_API_KEY": "fromfile"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPENAI_API_KEY", "")
	if err := Load(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("OPENAI_API_KEY"); got != "" {
		t.Errorf("an explicitly-cleared env var was overridden from file: %q", got)
	}
}

func TestLoadBlocksDisallowedAndEmptyKeys(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir, _ := ConfigDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path, _ := CredentialsPath()
	if err := os.WriteFile(path, []byte("=orphan\nLD_PRELOAD=/tmp/evil.so\nOPENAI_API_KEY=ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	os.Unsetenv("LD_PRELOAD")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("LD_PRELOAD")
	defer os.Unsetenv("OPENAI_API_KEY")

	if err := Load(); err != nil {
		t.Fatalf("Load must tolerate an empty-key line, got: %v", err)
	}
	if _, set := os.LookupEnv("LD_PRELOAD"); set {
		t.Error("LD_PRELOAD injected from credentials file — allowlist bypass (RCE vector)")
	}
	if os.Getenv("OPENAI_API_KEY") != "ok" {
		t.Error("allowlisted OPENAI_API_KEY not loaded")
	}
}

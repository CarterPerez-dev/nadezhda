// ©AngelaMos | 2026
// source_test.go

package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedDefaultsParse(t *testing.T) {
	got, err := Defaults()
	if err != nil {
		t.Fatalf("Defaults: %v", err)
	}
	want := map[string]bool{
		"krebs": false, "thehackernews": false, "bleepingcomputer": false,
		"securityweek": false, "darkreading": false, "theregister": false,
		"cisa": false,
	}
	for _, s := range got {
		if _, ok := want[s.Name]; !ok {
			continue
		}
		want[s.Name] = true
		if s.Type != KindRSS {
			t.Errorf("%s: type = %q, want rss", s.Name, s.Type)
		}
		if s.Weight <= 0 || s.Weight > 1 {
			t.Errorf("%s: weight %v out of (0,1]", s.Name, s.Weight)
		}
		if !s.Enabled {
			t.Errorf("%s: expected enabled by default", s.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("seed source %q missing from embedded defaults", name)
		}
	}
}

func TestExternalOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sources.yaml")
	body := `
- name: custom
  title: Custom Feed
  url: https://example.com/feed.xml
  type: rss
  weight: 0.5
  enabled: true
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 || got[0].Name != "custom" {
		t.Errorf("external override = %+v", got)
	}
}

func TestRejectsBadURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	body := "- name: x\n  url: not-a-url\n  type: rss\n  enabled: true\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for invalid url")
	}
}

func TestRejectsDuplicateName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dup.yaml")
	body := `
- {name: a, url: "https://a.com/f", type: rss, enabled: true}
- {name: a, url: "https://b.com/f", type: rss, enabled: true}
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error for duplicate source name")
	}
}

func TestEnabledFilter(t *testing.T) {
	all := []Source{
		{Name: "on", Enabled: true},
		{Name: "off", Enabled: false},
	}
	got := Enabled(all)
	if len(got) != 1 || got[0].Name != "on" {
		t.Errorf("Enabled filter = %+v", got)
	}
}

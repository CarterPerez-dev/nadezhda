// ©AngelaMos | 2026
// source.go

package source

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed sources.yaml
var embedded []byte

type Kind string

const (
	KindRSS  Kind = "rss"
	KindAtom Kind = "atom"
	KindHTML Kind = "html"
)

type Source struct {
	Name      string   `yaml:"name"`
	Title     string   `yaml:"title"`
	URL       string   `yaml:"url"`
	Type      Kind     `yaml:"type"`
	Extractor string   `yaml:"extractor"`
	Weight    float64  `yaml:"weight"`
	Tags      []string `yaml:"tags"`
	Enabled   bool     `yaml:"enabled"`
}

func Defaults() ([]Source, error) {
	return parse(embedded)
}

func Load(path string) ([]Source, error) {
	if path == "" {
		return Defaults()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults()
		}
		return nil, fmt.Errorf("read sources %s: %w", path, err)
	}
	return parse(raw)
}

func parse(raw []byte) ([]Source, error) {
	var out []Source
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse sources: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("sources: no entries")
	}
	seen := make(map[string]struct{}, len(out))
	for i, s := range out {
		if s.Name == "" {
			return nil, fmt.Errorf("sources[%d]: name is required", i)
		}
		if _, dup := seen[s.Name]; dup {
			return nil, fmt.Errorf("sources: duplicate name %q", s.Name)
		}
		seen[s.Name] = struct{}{}
		if _, err := url.ParseRequestURI(s.URL); err != nil {
			return nil, fmt.Errorf("sources[%s]: invalid url %q: %w", s.Name, s.URL, err)
		}
		switch s.Type {
		case KindRSS, KindAtom, KindHTML:
		default:
			return nil, fmt.Errorf("sources[%s]: type must be rss|atom|html, got %q", s.Name, s.Type)
		}
	}
	return out, nil
}

func Enabled(all []Source) []Source {
	out := make([]Source, 0, len(all))
	for _, s := range all {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

// ©AngelaMos | 2026
// wizard.go

package setup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	ProviderClaude = "anthropic"
	ProviderOpenAI = "openai"
	ProviderGemini = "gemini"
	ProviderQwen   = "qwen"

	envOpenAIKey    = "OPENAI_API_KEY"
	envGeminiKey    = "GEMINI_API_KEY"
	envAnthropicKey = "ANTHROPIC_API_KEY"
	envNVDKey       = "NVD_API_KEY"

	composeQwenURL = "http://localhost:39847/v1"
	nativeQwenURL  = "http://localhost:11434/v1"

	detectTimeout = 3 * time.Second
	ollamaInstall = "https://ollama.com/download"
)

type apiChoice struct {
	provider string
	envKey   string
	label    string
}

var apiProviders = map[string]apiChoice{
	"1": {ProviderClaude, envAnthropicKey, "Claude (Anthropic)"},
	"2": {ProviderOpenAI, envOpenAIKey, "OpenAI"},
	"3": {ProviderGemini, envGeminiKey, "Gemini (Google)"},
}

func Run(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, "nadezhda AI setup")
	fmt.Fprintln(out, "  1) Claude (Anthropic)   2) OpenAI   3) Gemini   4) Ollama (local qwen, free, no key)")
	choice := prompt(in, out, "choose a provider [1-4]: ")

	switch choice {
	case "1", "2", "3":
		return setupAPIProvider(in, out, apiProviders[choice])
	case "4":
		return setupOllama(out)
	default:
		return fmt.Errorf("setup: invalid choice %q (want 1-4)", choice)
	}
}

func setupAPIProvider(in io.Reader, out io.Writer, c apiChoice) error {
	fmt.Fprintf(out, "\nConfiguring %s.\n", c.label)
	key := strings.TrimSpace(readSecret(in, out, fmt.Sprintf("paste your %s API key: ", c.label)))
	if key == "" {
		return fmt.Errorf("setup: no API key entered")
	}
	updates := map[string]string{
		EnvProvider: c.provider,
		c.envKey:    key,
	}
	if nvd := strings.TrimSpace(readSecret(in, out, "optional NVD token for the CVE booster (enter to skip): ")); nvd != "" {
		updates[envNVDKey] = nvd
	}
	if err := Save(updates); err != nil {
		return err
	}
	path, _ := CredentialsPath()
	fmt.Fprintf(out, "\nSaved to %s (0600). %s is ready — run:  nadezhda ideate\n", path, c.label)
	return nil
}

func setupOllama(out io.Writer) error {
	fmt.Fprintln(out, "\nConfiguring Ollama (local qwen2.5, no API key).")
	client := &http.Client{Timeout: detectTimeout}

	if base := DetectOllama(client); base != "" {
		fmt.Fprintf(out, "Found a running Ollama at %s — using it.\n", base)
		return saveOllama(out, base+"/v1")
	}

	fmt.Fprintln(out, "No running Ollama found.")
	switch {
	case HasBinary("docker"):
		dir, _ := os.Getwd()
		path, err := WriteCompose(dir)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Wrote %s. Start it with:\n  docker compose -f %s up -d\n", path, path)
		fmt.Fprintln(out, "(first run pulls qwen2.5:7b, a few minutes), then re-run `nadezhda ai` and pick 4.")
		return saveOllama(out, composeQwenURL)
	case HasBinary("ollama"):
		fmt.Fprintln(out, "Ollama is installed but not running. Start it and pull the model:")
		fmt.Fprintln(out, "  ollama serve &   then:   ollama pull qwen2.5:7b")
		return saveOllama(out, nativeQwenURL)
	default:
		fmt.Fprintf(out, "Install Ollama (%s) or Docker, then re-run `nadezhda ai`.\n", ollamaInstall)
		return saveOllama(out, composeQwenURL)
	}
}

func saveOllama(out io.Writer, baseURL string) error {
	if err := Save(map[string]string{EnvProvider: ProviderQwen, EnvQwenURL: baseURL}); err != nil {
		return err
	}
	path, _ := CredentialsPath()
	fmt.Fprintf(out, "Saved qwen (%s) to %s. Run:  nadezhda ideate\n", baseURL, path)
	return nil
}

func prompt(in io.Reader, out io.Writer, label string) string {
	fmt.Fprint(out, label)
	return readLine(in)
}

func readSecret(in io.Reader, out io.Writer, label string) string {
	fmt.Fprint(out, label)
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(out)
		if err != nil {
			return ""
		}
		return string(b)
	}
	return readLine(in)
}

func readLine(in io.Reader) string {
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := in.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				break
			}
			if buf[0] != '\r' {
				b.WriteByte(buf[0])
			}
		}
		if err != nil {
			break
		}
	}
	return strings.TrimSpace(b.String())
}

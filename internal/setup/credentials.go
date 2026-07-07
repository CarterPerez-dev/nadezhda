// ©AngelaMos | 2026
// credentials.go

package setup

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	EnvProvider = "NADEZHDA_AI_PROVIDER"
	EnvQwenURL  = "NADEZHDA_QWEN_BASE_URL"

	commentMark = "#"
	kvSeparator = "="

	credentialKeyPrefix = "NADEZHDA_"
	credentialKeySuffix = "_API_KEY"
	tmpSuffix           = ".tmp"

	credentialsHeader = "# ©AngelaMos | 2026\n# nadezhda credentials — auto-managed by `nadezhda ai`; keep private, do not commit\n\n"
)

func Load() error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	kv, err := readCredentials(path)
	if err != nil {
		return err
	}
	for k, v := range kv {
		if !allowedKey(k) {
			continue
		}
		if _, set := os.LookupEnv(k); set {
			continue
		}
		if err := os.Setenv(k, v); err != nil {
			return fmt.Errorf("setup: set %s: %w", k, err)
		}
	}
	return nil
}

func allowedKey(k string) bool {
	return strings.HasPrefix(k, credentialKeyPrefix) || strings.HasSuffix(k, credentialKeySuffix)
}

func NonSecretEnviron() []string {
	env := os.Environ()
	out := make([]string, 0, len(env))
	for _, e := range env {
		if k, _, ok := strings.Cut(e, kvSeparator); ok && allowedKey(k) {
			continue
		}
		out = append(out, e)
	}
	return out
}

func Save(updates map[string]string) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("setup: create config dir: %w", err)
	}
	kv, err := readCredentials(path)
	if err != nil {
		return err
	}
	for k, v := range updates {
		kv[k] = v
	}
	return writeCredentials(path, kv)
}

func readCredentials(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("setup: read credentials: %w", err)
	}
	defer f.Close()

	kv := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, commentMark) {
			continue
		}
		k, v, ok := strings.Cut(line, kvSeparator)
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		kv[key] = strings.TrimSpace(v)
	}
	return kv, sc.Err()
}

func writeCredentials(path string, kv map[string]string) error {
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(credentialsHeader)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, kv[k])
	}

	tmp := path + tmpSuffix
	if err := os.WriteFile(tmp, []byte(b.String()), filePerm); err != nil {
		return fmt.Errorf("setup: write credentials: %w", err)
	}
	if err := os.Chmod(tmp, filePerm); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("setup: secure credentials: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("setup: replace credentials: %w", err)
	}
	return nil
}

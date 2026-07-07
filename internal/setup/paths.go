// ©AngelaMos | 2026
// paths.go

package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	appDir          = "nadezhda"
	credentialsFile = "credentials"
	dirPerm         = 0o700
	filePerm        = 0o600
)

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("setup: locate config dir: %w", err)
	}
	return filepath.Join(base, appDir), nil
}

func CredentialsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}

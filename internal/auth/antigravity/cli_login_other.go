//go:build !windows

package antigravity

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func readCLILoginPayload() ([]byte, error) {
	home, errHome := os.UserHomeDir()
	if errHome != nil {
		return nil, fmt.Errorf("resolve home directory: %w", errHome)
	}
	raw, errRead := os.ReadFile(filepath.Join(home, ".gemini", "antigravity-cli", "antigravity-oauth-token"))
	if errors.Is(errRead, fs.ErrNotExist) {
		return nil, ErrCLILoginNotFound
	}
	if errRead != nil {
		return nil, fmt.Errorf("read antigravity cli login: %w", errRead)
	}
	return raw, nil
}

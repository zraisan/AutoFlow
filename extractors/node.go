package extractors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mzkux/AutoFlow/types"
)

type nodeScripts struct {
	Scripts map[string]string `json:"scripts"`
}

func ReadNodeFiles(path string) (types.Scripts, error) {
	data, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return types.Scripts{}, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg nodeScripts
	if err := json.Unmarshal(data, &pkg); err != nil {
		return types.Scripts{}, fmt.Errorf("failed to parse package.json: %w", err)
	}

	return normalizeNodeScripts(pkg.Scripts), nil
}

func normalizeNodeScripts(raw map[string]string) types.Scripts {
	var s types.Scripts
	for key := range raw {
		keyLower := strings.ToLower(key)
		switch {
		case strings.Contains(keyLower, "lint"):
			s.Lint = "npm run " + key
		case strings.Contains(keyLower, "test"):
			s.Test = "npm run " + key
		case strings.Contains(keyLower, "build"):
			s.Build = "npm run " + key
		case strings.Contains(keyLower, "deploy"):
			s.Deploy = "npm run " + key
		}
	}
	s.Install = "npm install"
	return s
}

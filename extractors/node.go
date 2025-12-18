package extractors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type nodeScripts struct {
	Scripts map[string]string `json:"scripts"`
}

func ReadNodeFiles(path string) (Scripts, string, error) {
	data, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return Scripts{}, "", fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg nodeScripts
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Scripts{}, "", fmt.Errorf("failed to parse package.json: %w", err)
	}

	return normalizeNodeScripts(pkg.Scripts), path, nil
}

func normalizeNodeScripts(raw map[string]string) Scripts {
	var s Scripts
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

package extractors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zraisan/AutoFlow/registry"
)

func init() {
	registry.RegisterExtractor(&NodeExtractor{})
}

type NodeExtractor struct{}

type packageJSON struct {
	Scripts map[string]string `json:"scripts"`
	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`
}

func (n *NodeExtractor) Name() string {
	return "Node"
}

func (n *NodeExtractor) Extract(path string) (*registry.ExtractorResult, error) {
	data, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %w", err)
	}

	version, image := detectNodeVersion(path, pkg)
	result := &registry.ExtractorResult{
		Runtime:        "node",
		RuntimeVersion: version,
		Image:          image,
		PackageManager: detectNodePackageManager(path),
		Scripts:        normalizeNodeScripts(pkg.Scripts),
	}

	return result, nil
}

func detectNodeVersion(path string, pkg packageJSON) (string, string) {
	version := "20"

	if nvmrc, err := os.ReadFile(filepath.Join(path, ".nvmrc")); err == nil {
		v := strings.TrimSpace(string(nvmrc))
		v = strings.TrimPrefix(v, "v")
		if v != "" {
			version = v
		}
	} else if pkg.Engines.Node != "" {
		v := pkg.Engines.Node
		v = strings.TrimLeft(v, ">=^~")
		v = strings.Split(v, " ")[0]
		if v != "" {
			version = v
		}
	}

	return version, "node:" + version + "-alpine"
}

func detectNodePackageManager(path string) string {
	if _, err := os.Stat(filepath.Join(path, "pnpm-lock.yaml")); err == nil {
		return "pnpm"
	}
	if _, err := os.Stat(filepath.Join(path, "yarn.lock")); err == nil {
		return "yarn"
	}
	if _, err := os.Stat(filepath.Join(path, "bun.lockb")); err == nil {
		return "bun"
	}
	return "npm"
}

func normalizeNodeScripts(raw map[string]string) map[string]string {
	s := make(map[string]string)
	for key := range raw {
		keyLower := strings.ToLower(key)
		switch {
		case strings.Contains(keyLower, "lint"):
			s["Lint"] = "npm run " + key
		case strings.Contains(keyLower, "test"):
			s["Test"] = "npm run " + key
		case strings.Contains(keyLower, "build"):
			s["Build"] = "npm run " + key
		case strings.Contains(keyLower, "deploy"):
			s["Deploy"] = "npm run " + key
		}
	}
	s["Install"] = "npm install"
	return s
}

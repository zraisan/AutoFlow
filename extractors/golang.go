package extractors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mzkux/AutoFlow/registry"
)

func init() {
	registry.RegisterExtractor(&GolangExtractor{})
}

type GolangExtractor struct{}

func (g *GolangExtractor) Name() string {
	return "Golang"
}

func (g *GolangExtractor) Extract(path string) (*registry.ExtractorResult, error) {
	data, err := os.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	version, image := detectGoVersion(string(data))
	result := &registry.ExtractorResult{
		Runtime:        "go",
		RuntimeVersion: version,
		Image:          image,
		PackageManager: "go",
		Scripts: map[string]string{
			"Build": "go build -v ./...",
			"Test":  "go test -v ./...",
		},
	}

	return result, nil
}

func detectGoVersion(gomod string) (string, string) {
	for line := range strings.SplitSeq(gomod, "\n") {
		line = strings.TrimSpace(line)
		if after, found := strings.CutPrefix(line, "go "); found {
			version := strings.TrimSpace(after)
			if version >= "1.24" {
				return "1.24", "golang:latest"
			}
			return version, "golang:" + version
		}
	}
	return "latest", "golang:latest"
}

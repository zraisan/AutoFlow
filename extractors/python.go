package extractors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mzkux/AutoFlow/registry"
)

func init() {
	registry.RegisterExtractor(&PythonExtractor{})
}

type PythonExtractor struct{}

func (p *PythonExtractor) Name() string {
	return "Python"
}

func (p *PythonExtractor) Extract(path string) (*registry.ExtractorResult, error) {

	var packageManager string
	packageManager = detectPythonPackageManager(path)

	version, image := detectPythonVersion(path)
	result := &registry.ExtractorResult{
		Runtime:        "python",
		RuntimeVersion: version,
		Image:          image,
		PackageManager: packageManager,
		Scripts:        normalizePythonScripts(path, packageManager),
	}

	return result, nil
}

func detectPythonPackageManager(path string) string {
	data, err := os.ReadFile(filepath.Join(path, "pyproject.toml"))
	if err != nil {
		fmt.Println(err)
		return "pip"
	}

	for line := range strings.SplitAfterSeq(string(data), "\n") {
		if strings.Contains(line, "poetry") {
			return "poetry"
		}
	}

	return "uv"
}

func detectPythonVersion(path string) (string, string) {
	entries, err := os.ReadDir(filepath.Join(path, ".venv"))
	if err != nil {
		fmt.Println(err)
	}

	for _, entry := range entries {
		if after, found := strings.CutPrefix(entry.Name(), "python"); found {
			return after, "python:" + after + "-slim"
		}
	}
	return "3.13", "python:3.13-slim"
}

func normalizePythonScripts(path string, packageManager string) map[string]string {

	scripts := make(map[string]string)
	switch packageManager {
	case "pip":
		data, err := os.ReadFile(filepath.Join(path, "requirements.txt"))
		if err != nil {
			fmt.Println(err)
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(line, "ruff") {
				scripts["Lint"] = "ruff check ."
			}
			if strings.Contains(line, "pytest") {
				scripts["Test"] = "pytest"
			}

		}

	case "uv":
		data, err := os.ReadFile(filepath.Join(path, "pyproject.toml"))
		if err != nil {
			fmt.Println(err)
		}

		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(line, "ruff") {
				scripts["Lint"] = "uv run ruff check ."
			}
			if strings.Contains(line, "pytest") {
				scripts["Test"] = "uv run pytest"
			}
		}

	case "poetry":
		data, err := os.ReadFile(filepath.Join(path, "pyproject.toml"))
		if err != nil {
			fmt.Println(err)
		}

		for line := range strings.SplitSeq(string(data), "\n") {
			if strings.Contains(line, "ruff") {
				scripts["Lint"] = "poetry run ruff check ."
			}
			if strings.Contains(line, "pytest") {
				scripts["Test"] = "poetry run pytest"
			}
		}
	}

	return scripts
}

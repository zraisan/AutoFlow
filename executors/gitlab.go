package executors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zraisan/AutoFlow/registry"
	"gopkg.in/yaml.v3"
)

func init() {
	registry.RegisterExecutor(&GitlabExecutor{})
}

type GitlabExecutor struct{}

type gitlabWorkflow struct {
	Stages []string             `yaml:"stages"`
	Jobs   map[string]gitlabJob `yaml:",inline"`
}

type gitlabJob struct {
	Stage     string          `yaml:"stage"`
	Image     string          `yaml:"image,omitempty"`
	Script    []string        `yaml:"script"`
	Only      []string        `yaml:"only,omitempty"`
	Artifacts gitlabArtifacts `yaml:"artifacts,omitempty"`
}

type gitlabArtifacts struct {
	Paths []string `yaml:"paths,omitempty"`
}

func (g *GitlabExecutor) Name() string {
	return "Gitlab"
}

func (g *GitlabExecutor) Generate(result *registry.ExtractorResult, path, name string) (string, error) {
	var stages []string
	jobs := make(map[string]gitlabJob)

	for key, value := range result.Scripts {
		if key == "Install" {
			continue
		}
		stages = append(stages, strings.ToLower(key))

		var script []string
		if result.Scripts["Install"] != "" {
			script = append(script, result.Scripts["Install"])
		}
		script = append(script, value)

		jobs[key] = gitlabJob{
			Stage:  strings.ToLower(key),
			Image:  result.Image,
			Script: script,
		}
	}

	workflow := &gitlabWorkflow{
		Stages: stages,
		Jobs:   jobs,
	}

	data, err := yaml.Marshal(&workflow)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow: %w", err)
	}

	workflowFilepath := filepath.Join(path, fmt.Sprintf(".gitlab-ci-%s.yml", name))
	if err := os.WriteFile(workflowFilepath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write workflow file: %w", err)
	}

	return string(data), nil
}

package executors

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/zraisan/AutoFlow/registry"
	"gopkg.in/yaml.v3"
)

func init() {
	registry.RegisterExecutor(&GithubExecutor{})
}

type GithubExecutor struct{}

type githubWorkflow struct {
	Name string               `yaml:"name"`
	On   githubOn             `yaml:"on"`
	Jobs map[string]githubJob `yaml:"jobs"`
}

type githubOn struct {
	Push        *githubPushEvent        `yaml:"push,omitempty"`
	PullRequest *githubPullRequestEvent `yaml:"pull_request,omitempty"`
}

type githubPushEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type githubPullRequestEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type githubJob struct {
	RunsOn string       `yaml:"runs-on"`
	Steps  []githubStep `yaml:"steps"`
}

type githubStep struct {
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
}

func (g *GithubExecutor) Name() string {
	return "GitHub"
}

func (g *GithubExecutor) Generate(result *registry.ExtractorResult, path, name string) (string, error) {
	steps := []githubStep{
		{Name: "Checkout", Uses: "actions/checkout@v4"},
	}

	setupSteps := g.createSetupSteps(result)
	steps = append(steps, setupSteps...)

	for stepName, command := range result.Scripts {
		steps = append(steps, githubStep{
			Name: stepName,
			Run:  command,
		})
	}

	workflow := githubWorkflow{
		Name: name,
		On: githubOn{
			Push: &githubPushEvent{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]githubJob{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  steps,
			},
		},
	}

	data, err := yaml.Marshal(&workflow)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow: %w", err)
	}

	workflowDir := filepath.Join(path, ".github/workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		log.Fatal(err)
	}

	if len(name) == 0 {
		name = "ci"
	}
	workflowPath := filepath.Join(workflowDir, fmt.Sprintf("%s.yml", name))
	if err := os.WriteFile(workflowPath, data, 0644); err != nil {
		log.Fatal(err)
	}

	return string(data), nil
}

func (g *GithubExecutor) createSetupSteps(result *registry.ExtractorResult) []githubStep {
	switch result.Runtime {
	case "node":
		return []githubStep{
			{
				Name: "Setup Node.js",
				Uses: "actions/setup-node@v4",
				With: map[string]string{
					"node-version": result.RuntimeVersion,
				},
			},
		}
	case "go":
		return []githubStep{
			{
				Name: "Setup Go",
				Uses: "actions/setup-go@v5",
				With: map[string]string{
					"go-version": result.RuntimeVersion,
				},
			},
			{
				Name: "Golangci-lint",
				Uses: "golangci/golangci-lint-action@v7",
			},
		}
	case "python":
		steps := []githubStep{
			{
				Name: "Setup Python",
				Uses: "actions/setup-python@v5",
				With: map[string]string{
					"python-version": result.RuntimeVersion,
				},
			},
		}

		switch result.PackageManager {
		case "uv":
			steps = append(steps, githubStep{
				Name: "Install uv",
				Uses: "astral-sh/setup-uv@v4",
			})
		case "poetry":
			steps = append(steps, githubStep{
				Name: "Install Poetry",
				Uses: "snok/install-poetry@v1",
			})
		}
		return steps
	}
	return nil
}

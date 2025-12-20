package executors

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mzkux/AutoFlow/types"
	"gopkg.in/yaml.v3"
)

type GithubWorkflow struct {
	Name string               `yaml:"name"`
	On   GithubOn             `yaml:"on"`
	Jobs map[string]GithubJob `yaml:"jobs"`
}

type GithubOn struct {
	Push        *GithubPushEvent        `yaml:"push,omitempty"`
	PullRequest *GithubPullRequestEvent `yaml:"pull_request,omitempty"`
}

type GithubPushEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type GithubPullRequestEvent struct {
	Branches []string `yaml:"branches,omitempty"`
}

type GithubJob struct {
	RunsOn string       `yaml:"runs-on"`
	Steps  []GithubStep `yaml:"steps"`
}

type GithubStep struct {
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
}

func WriteGithubYaml(scripts types.Scripts, path string, name string) string {
	steps := []GithubStep{
		{Name: "Checkout", Uses: "actions/checkout@v4"},
	}

	if scripts.Install != "" {
		steps = append(steps, GithubStep{Name: "Install", Run: scripts.Install})
	}
	if scripts.Lint != "" {
		steps = append(steps, GithubStep{Name: "Lint", Run: scripts.Lint})
	}
	if scripts.Build != "" {
		steps = append(steps, GithubStep{Name: "Build", Run: scripts.Build})
	}
	if scripts.Test != "" {
		steps = append(steps, GithubStep{Name: "Test", Run: scripts.Test})
	}
	if scripts.Deploy != "" {
		steps = append(steps, GithubStep{Name: "Deploy", Run: scripts.Deploy})
	}

	dataz, err := yaml.Marshal(&GithubWorkflow{
		Name: name,
		On: GithubOn{
			Push: &GithubPushEvent{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]GithubJob{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  steps,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(path, ".github/workflows"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, fmt.Sprintf(".github/workflows/%s.yml", name)), dataz, 0644); err != nil {
		log.Fatal(err)
	}

	return string(dataz)
}

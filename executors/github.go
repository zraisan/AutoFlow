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

func WriteGithubYaml(scripts types.Scripts, path string) string {
	dataz, err := yaml.Marshal(&GithubWorkflow{
		Name: "test",
		On: GithubOn{
			Push: &GithubPushEvent{
				Branches: []string{"main"},
			},
		},
		Jobs: map[string]GithubJob{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []GithubStep{
					{Name: "Checkout", Uses: "actions/checkout@v4"},
					{Name: "Install", Run: scripts.Install},
					{Name: "Lint", Run: scripts.Lint},
					{Name: "Build", Run: scripts.Build},
					{Name: "Test", Run: scripts.Test},
					{Name: "Deploy", Run: scripts.Deploy},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(filepath.Join(path, ".github/workflows"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, ".github/workflows/node.yml"), dataz, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Println(filepath.Join(path, ".github/workflows"))
	fmt.Println(string(dataz))
	return string(dataz)
}

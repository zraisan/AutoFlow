package executors

import (
	"strings"
	"testing"

	"github.com/mzkux/AutoFlow/extractors"
	"gopkg.in/yaml.v3"
)

func TestWriteGithubYaml(t *testing.T) {
	scripts := extractors.Scripts{
		Install: "npm install",
		Lint:    "npm run lint",
		Build:   "npm run build",
		Test:    "npm run test",
	}
	result := WriteGithubYaml(scripts)

	if result == "" {
		t.Fatal("WriteGithubYaml returned empty string")
	}

	var workflow GithubWorkflow
	if err := yaml.Unmarshal([]byte(result), &workflow); err != nil {
		t.Fatalf("WriteGithubYaml produced invalid YAML: %v", err)
	}

	if workflow.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", workflow.Name)
	}

	if workflow.On.Push == nil {
		t.Fatal("expected push event, got nil")
	}
	if len(workflow.On.Push.Branches) != 1 || workflow.On.Push.Branches[0] != "main" {
		t.Errorf("expected push branches ['main'], got %v", workflow.On.Push.Branches)
	}

	job, ok := workflow.Jobs["build"]
	if !ok {
		t.Fatal("expected 'build' job")
	}

	if job.RunsOn != "ubuntu-latest" {
		t.Errorf("expected runs-on 'ubuntu-latest', got '%s'", job.RunsOn)
	}

	if len(job.Steps) != 6 {
		t.Errorf("expected 6 steps, got %d", len(job.Steps))
	}
}

func TestWriteGithubYamlContainsExpectedStrings(t *testing.T) {
	scripts := extractors.Scripts{
		Install: "npm install",
		Lint:    "npm run lint",
		Build:   "npm run build",
		Test:    "npm run test",
	}
	result := WriteGithubYaml(scripts)

	expectedStrings := []string{
		"name: test",
		"runs-on: ubuntu-latest",
		"actions/checkout@v4",
		"npm install",
		"npm run test",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain '%s'", expected)
		}
	}
}

func TestGithubWorkflowStructMarshal(t *testing.T) {
	w := GithubWorkflow{
		Name: "custom",
		On: GithubOn{
			PullRequest: &GithubPullRequestEvent{
				Branches: []string{"develop", "main"},
			},
		},
		Jobs: map[string]GithubJob{
			"test": {
				RunsOn: "ubuntu-22.04",
				Steps: []GithubStep{
					{Name: "Run tests", Run: "go test ./..."},
				},
			},
		},
	}

	data, err := yaml.Marshal(&w)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got GithubWorkflow
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Name != "custom" {
		t.Errorf("expected name 'custom', got '%s'", got.Name)
	}

	if got.On.PullRequest == nil {
		t.Fatal("expected pull_request event")
	}

	if len(got.On.PullRequest.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(got.On.PullRequest.Branches))
	}
}

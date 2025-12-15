package executors

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWriteYaml(t *testing.T) {
	result := WriteYaml("")

	if result == "" {
		t.Fatal("WriteYaml returned empty string")
	}

	var workflow Workflow
	if err := yaml.Unmarshal([]byte(result), &workflow); err != nil {
		t.Fatalf("WriteYaml produced invalid YAML: %v", err)
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

	if len(job.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(job.Steps))
	}
}

func TestWriteYamlContainsExpectedStrings(t *testing.T) {
	result := WriteYaml("")

	expectedStrings := []string{
		"name: test",
		"runs-on: ubuntu-latest",
		"actions/checkout@v4",
		"npm install",
		"npm test",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("expected output to contain '%s'", expected)
		}
	}
}

func TestWorkflowStructMarshal(t *testing.T) {
	w := Workflow{
		Name: "custom",
		On: On{
			PullRequest: &PullRequestEvent{
				Branches: []string{"develop", "main"},
			},
		},
		Jobs: map[string]Job{
			"test": {
				RunsOn: "ubuntu-22.04",
				Steps: []Step{
					{Name: "Run tests", Run: "go test ./..."},
				},
			},
		},
	}

	data, err := yaml.Marshal(&w)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got Workflow
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

package extractors

import (
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	result := readFiles("../data/node")

	if result == "" {
		t.Fatal("returned empty string")
	}

	if !strings.HasPrefix(strings.TrimSpace(result), "{") {
		t.Errorf("expected JSON object, got %q", result)
	}

	if !strings.Contains(result, "name") {
		t.Error("expected package.json to contain 'name' field")
	}
}

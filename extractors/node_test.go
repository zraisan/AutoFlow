package extractors

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	scripts, err := ReadNodeFiles("../data/node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scripts.Install != "npm install" {
		t.Errorf("expected Install to be 'npm install', got %q", scripts.Install)
	}
}

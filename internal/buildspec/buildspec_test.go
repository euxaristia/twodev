package buildspec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRepoBuildSpec(t *testing.T) {
	path := filepath.Join("..", "..", ".onedev-buildspec.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	spec, err := Parse(string(content))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Version != 47 {
		t.Fatalf("version = %d, want 47", spec.Version)
	}
	if len(spec.Jobs) < 5 {
		t.Fatalf("expected at least 5 jobs, got %d", len(spec.Jobs))
	}

	summary := Summarize(spec)
	if summary.JobCount != len(spec.Jobs) {
		t.Fatalf("summary job count = %d, want %d", summary.JobCount, len(spec.Jobs))
	}
	foundCI := false
	for _, name := range summary.JobNames {
		if name == "CI" {
			foundCI = true
			break
		}
	}
	if !foundCI {
		t.Fatal("expected CI job in build spec")
	}
}

func TestParseRejectsDuplicateJobs(t *testing.T) {
	content := `
version: 1
jobs:
- name: build
  steps:
  - type: CommandStep
    name: run
    interpreter:
      type: DefaultInterpreter
      commands: echo hi
- name: build
  steps:
  - type: CommandStep
    name: run
    interpreter:
      type: DefaultInterpreter
      commands: echo hi
`
	if _, err := Parse(content); err == nil {
		t.Fatal("expected duplicate job error")
	}
}
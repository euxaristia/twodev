package buildspec

import "testing"

func TestJobsForBranchUpdate(t *testing.T) {
	spec, err := Parse(`
version: 1
jobs:
- name: CI
  steps:
  - type: CommandStep
    name: build
    interpreter:
      type: DefaultInterpreter
      commands: echo ci
  triggers:
  - type: BranchUpdateTrigger
    branches: main
    projects: onedev/server
`)
	if err != nil {
		t.Fatal(err)
	}
	names := JobsForBranchUpdate(spec, "main", "onedev/server")
	if len(names) != 1 || names[0] != "CI" {
		t.Fatalf("jobs = %v", names)
	}
	if len(JobsForBranchUpdate(spec, "dev", "onedev/server")) != 0 {
		t.Fatal("expected no jobs for dev branch")
	}
}
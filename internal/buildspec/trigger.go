package buildspec

import "strings"

// MatchBranchUpdate reports whether a BranchUpdateTrigger fires for branch and project.
func MatchBranchUpdate(job Job, branch, project string) bool {
	for _, trigger := range job.Triggers {
		if trigger.Type != "BranchUpdateTrigger" {
			continue
		}
		branches := stringListField(trigger.Fields, "branches")
		projects := stringListField(trigger.Fields, "projects")
		if len(branches) > 0 && !contains(branches, branch) {
			continue
		}
		if len(projects) > 0 && !contains(projects, project) {
			continue
		}
		return true
	}
	return false
}

// JobsForBranchUpdate returns job names that should run for a branch update.
func JobsForBranchUpdate(spec *BuildSpec, branch, project string) []string {
	var names []string
	for _, job := range spec.Jobs {
		if MatchBranchUpdate(job, branch, project) {
			names = append(names, job.Name)
		}
	}
	return names
}

func stringListField(fields map[string]any, key string) []string {
	raw, ok := fields[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case string:
		return []string{v}
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}
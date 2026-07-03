package buildspec

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse decodes YAML build spec content into a BuildSpec.
func Parse(content string) (*BuildSpec, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil, fmt.Errorf("expected yaml document")
	}

	spec, err := decodeBuildSpec(root.Content[0])
	if err != nil {
		return nil, err
	}
	if err := validate(spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// Summarize returns a compact overview of a build spec.
func Summarize(spec *BuildSpec) Summary {
	summary := Summary{Version: spec.Version}
	for _, job := range spec.Jobs {
		summary.JobNames = append(summary.JobNames, job.Name)
	}
	for _, template := range spec.StepTemplates {
		summary.TemplateNames = append(summary.TemplateNames, template.Name)
	}
	summary.JobCount = len(summary.JobNames)
	summary.TemplateCount = len(summary.TemplateNames)
	return summary
}

func decodeBuildSpec(node *yaml.Node) (*BuildSpec, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("build spec root must be a mapping")
	}

	spec := &BuildSpec{}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		switch key {
		case "version":
			var version int
			if err := value.Decode(&version); err != nil {
				return nil, fmt.Errorf("decode version: %w", err)
			}
			spec.Version = version
		case "imports":
			if err := value.Decode(&spec.Imports); err != nil {
				return nil, fmt.Errorf("decode imports: %w", err)
			}
		case "jobs":
			jobs, err := decodeJobs(value)
			if err != nil {
				return nil, err
			}
			spec.Jobs = jobs
		case "stepTemplates":
			templates, err := decodeStepTemplates(value)
			if err != nil {
				return nil, err
			}
			spec.StepTemplates = templates
		case "services":
			if err := value.Decode(&spec.Services); err != nil {
				return nil, fmt.Errorf("decode services: %w", err)
			}
		case "properties":
			if err := value.Decode(&spec.Properties); err != nil {
				return nil, fmt.Errorf("decode properties: %w", err)
			}
		}
	}
	return spec, nil
}

func decodeJobs(node *yaml.Node) ([]Job, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("jobs must be a sequence")
	}
	jobs := make([]Job, 0, len(node.Content))
	for _, item := range node.Content {
		job, err := decodeJob(item)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func decodeJob(node *yaml.Node) (Job, error) {
	if node.Kind != yaml.MappingNode {
		return Job{}, fmt.Errorf("job must be a mapping")
	}

	job := Job{}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		switch key {
		case "name":
			job.Name = scalarValue(value)
		case "jobExecutor":
			job.JobExecutor = scalarValue(value)
		case "steps":
			steps, err := decodeSteps(value)
			if err != nil {
				return Job{}, err
			}
			job.Steps = steps
		case "paramSpecs":
			if err := value.Decode(&job.ParamSpecs); err != nil {
				return Job{}, fmt.Errorf("decode paramSpecs: %w", err)
			}
		case "jobDependencies":
			if err := value.Decode(&job.JobDependencies); err != nil {
				return Job{}, fmt.Errorf("decode jobDependencies: %w", err)
			}
		case "projectDependencies":
			if err := value.Decode(&job.ProjectDependencies); err != nil {
				return Job{}, fmt.Errorf("decode projectDependencies: %w", err)
			}
		case "requiredServices":
			if err := value.Decode(&job.RequiredServices); err != nil {
				return Job{}, fmt.Errorf("decode requiredServices: %w", err)
			}
		case "triggers":
			triggers, err := decodeTriggers(value)
			if err != nil {
				return Job{}, err
			}
			job.Triggers = triggers
		case "postBuildActions":
			if err := value.Decode(&job.PostBuildActions); err != nil {
				return Job{}, fmt.Errorf("decode postBuildActions: %w", err)
			}
		case "retryCondition":
			job.RetryCondition = scalarValue(value)
		case "maxRetries":
			var retries int
			if err := value.Decode(&retries); err != nil {
				return Job{}, fmt.Errorf("decode maxRetries: %w", err)
			}
			job.MaxRetries = retries
		case "retryDelay":
			var delay int
			if err := value.Decode(&delay); err != nil {
				return Job{}, fmt.Errorf("decode retryDelay: %w", err)
			}
			job.RetryDelay = delay
		case "timeout":
			var timeout int
			if err := value.Decode(&timeout); err != nil {
				return Job{}, fmt.Errorf("decode timeout: %w", err)
			}
			job.Timeout = timeout
		}
	}
	return job, nil
}

func decodeStepTemplates(node *yaml.Node) ([]StepTemplate, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("stepTemplates must be a sequence")
	}
	templates := make([]StepTemplate, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("step template must be a mapping")
		}
		template := StepTemplate{}
		for i := 0; i < len(item.Content); i += 2 {
			key := item.Content[i].Value
			value := item.Content[i+1]
			switch key {
			case "name":
				template.Name = scalarValue(value)
			case "steps":
				steps, err := decodeSteps(value)
				if err != nil {
					return nil, err
				}
				template.Steps = steps
			case "paramSpecs":
				if err := value.Decode(&template.ParamSpecs); err != nil {
					return nil, fmt.Errorf("decode paramSpecs: %w", err)
				}
			}
		}
		templates = append(templates, template)
	}
	return templates, nil
}

func decodeSteps(node *yaml.Node) ([]Step, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("steps must be a sequence")
	}
	steps := make([]Step, 0, len(node.Content))
	for _, item := range node.Content {
		step, err := decodeStep(item)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func decodeStep(node *yaml.Node) (Step, error) {
	if node.Kind != yaml.MappingNode {
		return Step{}, fmt.Errorf("step must be a mapping")
	}

	step := Step{Fields: make(map[string]any)}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		switch key {
		case "type":
			step.Type = scalarValue(value)
		case "name":
			step.Name = scalarValue(value)
		case "condition":
			step.Condition = scalarValue(value)
		case "optional":
			var optional bool
			if err := value.Decode(&optional); err != nil {
				return Step{}, fmt.Errorf("decode optional: %w", err)
			}
			step.Optional = optional
		default:
			var decoded any
			if err := value.Decode(&decoded); err != nil {
				return Step{}, fmt.Errorf("decode step field %q: %w", key, err)
			}
			step.Fields[key] = decoded
		}
	}
	return step, nil
}

func decodeTriggers(node *yaml.Node) ([]Trigger, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("triggers must be a sequence")
	}
	triggers := make([]Trigger, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("trigger must be a mapping")
		}
		trigger := Trigger{Fields: make(map[string]any)}
		for i := 0; i < len(item.Content); i += 2 {
			key := item.Content[i].Value
			value := item.Content[i+1]
			if key == "type" {
				trigger.Type = scalarValue(value)
				continue
			}
			var decoded any
			if err := value.Decode(&decoded); err != nil {
				return nil, fmt.Errorf("decode trigger field %q: %w", key, err)
			}
			trigger.Fields[key] = decoded
		}
		triggers = append(triggers, trigger)
	}
	return triggers, nil
}

func scalarValue(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Value)
}

func validate(spec *BuildSpec) error {
	if spec.Version <= 0 {
		return fmt.Errorf("version must be positive")
	}
	seenJobs := make(map[string]struct{}, len(spec.Jobs))
	for _, job := range spec.Jobs {
		if strings.TrimSpace(job.Name) == "" {
			return fmt.Errorf("job name is required")
		}
		if _, exists := seenJobs[job.Name]; exists {
			return fmt.Errorf("duplicate job name %q", job.Name)
		}
		seenJobs[job.Name] = struct{}{}
		if len(job.Steps) == 0 {
			return fmt.Errorf("job %q must define at least one step", job.Name)
		}
		for _, step := range job.Steps {
			if err := validateStep(job.Name, step); err != nil {
				return err
			}
		}
	}
	seenTemplates := make(map[string]struct{}, len(spec.StepTemplates))
	for _, template := range spec.StepTemplates {
		if strings.TrimSpace(template.Name) == "" {
			return fmt.Errorf("step template name is required")
		}
		if _, exists := seenTemplates[template.Name]; exists {
			return fmt.Errorf("duplicate step template name %q", template.Name)
		}
		seenTemplates[template.Name] = struct{}{}
	}
	return nil
}

func validateStep(jobName string, step Step) error {
	if strings.TrimSpace(step.Type) == "" {
		return fmt.Errorf("job %q has step without type", jobName)
	}
	if strings.TrimSpace(step.Name) == "" {
		return fmt.Errorf("job %q step %q requires name", jobName, step.Type)
	}
	return nil
}
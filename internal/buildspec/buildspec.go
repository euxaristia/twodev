package buildspec

const BlobPath = ".onedev-buildspec.yml"

// BuildSpec is the top-level CI/CD definition stored in .onedev-buildspec.yml.
type BuildSpec struct {
	Version       int           `yaml:"version"`
	Imports       []Import      `yaml:"imports"`
	Jobs          []Job         `yaml:"jobs"`
	StepTemplates []StepTemplate `yaml:"stepTemplates"`
	Services      []Service     `yaml:"services"`
	Properties    []Property    `yaml:"properties"`
}

// Import references another project's build spec.
type Import struct {
	ProjectPath       string `yaml:"projectPath"`
	Revision          string `yaml:"revision"`
	AccessTokenSecret string `yaml:"accessTokenSecret"`
}

// Job is a named pipeline with steps and optional triggers.
type Job struct {
	Name              string           `yaml:"name"`
	JobExecutor       string           `yaml:"jobExecutor"`
	Steps             []Step           `yaml:"steps"`
	ParamSpecs        []map[string]any `yaml:"paramSpecs"`
	JobDependencies   []JobDependency  `yaml:"jobDependencies"`
	ProjectDependencies []map[string]any `yaml:"projectDependencies"`
	RequiredServices  []string         `yaml:"requiredServices"`
	Triggers          []Trigger        `yaml:"triggers"`
	PostBuildActions  []map[string]any `yaml:"postBuildActions"`
	RetryCondition    string           `yaml:"retryCondition"`
	MaxRetries        int              `yaml:"maxRetries"`
	RetryDelay        int              `yaml:"retryDelay"`
	Timeout           int              `yaml:"timeout"`
}

// JobDependency links jobs via artifacts.
type JobDependency struct {
	JobName           string `yaml:"jobName"`
	RequireSuccessful bool   `yaml:"requireSuccessful"`
	Artifacts         string `yaml:"artifacts"`
}

// StepTemplate is a reusable step group.
type StepTemplate struct {
	Name       string           `yaml:"name"`
	Steps      []Step           `yaml:"steps"`
	ParamSpecs []map[string]any `yaml:"paramSpecs"`
}

// Service describes a long-running service used by jobs.
type Service struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// Property is a job-level property definition.
type Property struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// Step is a polymorphic pipeline step. The Type field mirrors the Java @Editable name.
type Step struct {
	Type      string         `yaml:"type"`
	Name      string         `yaml:"name"`
	Condition string         `yaml:"condition"`
	Optional  bool           `yaml:"optional"`
	Fields    map[string]any `yaml:",inline"`
}

// Trigger is a polymorphic job trigger.
type Trigger struct {
	Type   string         `yaml:"type"`
	Fields map[string]any `yaml:",inline"`
}

// Summary is a lightweight view used by APIs and tooling.
type Summary struct {
	Version       int
	JobCount      int
	JobNames      []string
	TemplateCount int
	TemplateNames []string
}
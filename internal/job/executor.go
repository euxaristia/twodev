package job

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
)

// Executor runs buildspec jobs on the local machine.
type Executor struct {
	workRoot string
	repoRoot string
	git      *git.Service
	logger   *Logger
}

// NewExecutor creates a local job executor rooted at workRoot.
func NewExecutor(workRoot string, logger *Logger) *Executor {
	return &Executor{workRoot: workRoot, git: git.NewService(""), logger: logger}
}

// NewExecutorWithRepo creates an executor that can run checkout steps.
func NewExecutorWithRepo(workRoot, repoRoot string, logger *Logger) *Executor {
	return &Executor{workRoot: workRoot, repoRoot: repoRoot, git: git.NewService(""), logger: logger}
}

// RunJob executes all steps for a named job from a parsed build spec.
func (e *Executor) RunJob(ctx context.Context, spec *buildspec.BuildSpec, jobName string, jobCtx Context) error {
	job, ok := findJob(spec, jobName)
	if !ok {
		return fmt.Errorf("job %q not found", jobName)
	}

	workDir := jobCtx.BuildDir(e.workRoot)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return err
	}
	jobCtx.WorkDir = workDir
	if jobCtx.RepoRoot == "" {
		jobCtx.RepoRoot = e.repoRoot
	}

	for _, step := range job.Steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := e.runStep(ctx, jobCtx, step); err != nil {
			if step.Optional {
				e.logger.Log(fmt.Sprintf("optional step %q failed: %v", step.Name, err))
				continue
			}
			return fmt.Errorf("step %q: %w", step.Name, err)
		}
	}
	return nil
}

func (e *Executor) runStep(ctx context.Context, jobCtx Context, step buildspec.Step) error {
	e.logger.Log(fmt.Sprintf("running step %q (%s)", step.Name, step.Type))
	switch step.Type {
	case "CommandStep":
		return e.runCommandStep(ctx, jobCtx, step)
	case "CheckoutStep":
		return e.runCheckoutStep(ctx, jobCtx, step)
	default:
		e.logger.Log(fmt.Sprintf("skipping unsupported step type %s", step.Type))
		return nil
	}
}

func (e *Executor) runCheckoutStep(ctx context.Context, jobCtx Context, step buildspec.Step) error {
	cloneURL, err := e.checkoutSource(jobCtx)
	if err != nil {
		return err
	}
	checkoutPath := "work"
	if raw, ok := step.Fields["checkoutPath"].(string); ok && strings.TrimSpace(raw) != "" {
		checkoutPath = strings.TrimSpace(raw)
	}
	dest := filepath.Join(jobCtx.WorkDir, checkoutPath)

	depth := 0
	if raw, ok := step.Fields["cloneDepth"].(int); ok && raw > 0 {
		depth = raw
	}
	if raw, ok := step.Fields["cloneDepth"].(float64); ok && int(raw) > 0 {
		depth = int(raw)
	}
	withLfs, _ := step.Fields["withLfs"].(bool)
	withSubmodules, _ := step.Fields["withSubmodules"].(bool)

	e.logger.Log(fmt.Sprintf("cloning %s into %s", jobCtx.ProjectPath, dest))
	if err := e.git.Clone(ctx, git.CloneOptions{
		URL:            cloneURL,
		Branch:         jobCtx.Branch,
		Depth:          depth,
		WithLFS:        withLfs,
		WithSubmodules: withSubmodules,
		DestDir:        dest,
	}); err != nil {
		return err
	}
	if jobCtx.CommitHash != "" {
		return e.git.Checkout(ctx, dest, jobCtx.CommitHash)
	}
	return nil
}

func (e *Executor) checkoutSource(jobCtx Context) (string, error) {
	if jobCtx.CloneURL != "" {
		return jobCtx.CloneURL, nil
	}
	repoRoot := jobCtx.RepoRoot
	if repoRoot == "" {
		repoRoot = e.repoRoot
	}
	if repoRoot == "" {
		return "", fmt.Errorf("checkout requires repo root or clone URL")
	}
	repoDir := filepath.Join(repoRoot, jobCtx.ProjectPath+".git")
	if _, err := os.Stat(repoDir); err != nil {
		return "", fmt.Errorf("local repository not found at %s", repoDir)
	}
	return repoDir, nil
}

func (e *Executor) runCommandStep(ctx context.Context, jobCtx Context, step buildspec.Step) error {
	commands, err := commandLines(step)
	if err != nil {
		return err
	}
	shell, flag := shellForOS()
	script := strings.Join(commands, "\n")
	cmd := exec.CommandContext(ctx, shell, flag, script)
	cmd.Dir = filepath.Join(jobCtx.WorkDir, "work")
	cmd.Env = os.Environ()
	cmd.Stdout = writerFunc(func(p []byte) (int, error) {
		e.logger.Log(strings.TrimRight(string(p), "\r\n"))
		return len(p), nil
	})
	cmd.Stderr = writerFunc(func(p []byte) (int, error) {
		e.logger.Log(strings.TrimRight(string(p), "\r\n"))
		return len(p), nil
	})
	if err := os.MkdirAll(cmd.Dir, 0o755); err != nil {
		return err
	}
	return cmd.Run()
}

func commandLines(step buildspec.Step) ([]string, error) {
	interpreter, ok := step.Fields["interpreter"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("command step missing interpreter")
	}
	raw, ok := interpreter["commands"]
	if !ok {
		return nil, fmt.Errorf("command step missing commands")
	}
	switch v := raw.(type) {
	case string:
		lines := strings.Split(strings.ReplaceAll(v, "\r\n", "\n"), "\n")
		out := make([]string, 0, len(lines))
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				out = append(out, line)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported commands type %T", raw)
	}
}

func shellForOS() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "sh", "-ec"
}

func findJob(spec *buildspec.BuildSpec, name string) (buildspec.Job, bool) {
	for _, job := range spec.Jobs {
		if job.Name == name {
			return job, true
		}
	}
	return buildspec.Job{}, false
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
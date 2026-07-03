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
)

// Executor runs buildspec jobs on the local machine.
type Executor struct {
	workRoot string
	logger   *Logger
}

// NewExecutor creates a local job executor rooted at workRoot.
func NewExecutor(workRoot string, logger *Logger) *Executor {
	return &Executor{workRoot: workRoot, logger: logger}
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
		e.logger.Log("checkout step delegated to git service in a later slice")
		return nil
	default:
		e.logger.Log(fmt.Sprintf("skipping unsupported step type %s", step.Type))
		return nil
	}
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
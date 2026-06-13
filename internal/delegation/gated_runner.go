package delegation

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/blastradius"
	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
)

// runner is an unexported subprocess manager. Only GatedRunner can
// create and use one — structural enforcement that every spawn goes
// through the Gatekeeper.
type runner struct {
	cfg    AgentConfig
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
}

func newRunner(cfg AgentConfig) *runner {
	return &runner{cfg: cfg}
}

func (r *runner) start(ctx context.Context, req *SpawnRequest) error {
	args := r.buildArgs(req)
	r.cmd = exec.CommandContext(ctx, r.cfg.Command, args...) //nolint:gosec // CLI is user-installed, not arbitrary
	stdinPipe, err := r.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("delegation: stdin: %w", err)
	}
	// Always close stdin on error paths; overridden on success below.
	defer func() { _ = stdinPipe.Close() }()

	stdoutPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("delegation: stdout: %w", err)
	}
	r.stdin = bufio.NewWriter(stdinPipe)
	r.stdout = bufio.NewReader(stdoutPipe)
	// Write the task to stdin.
	if _, err := r.stdin.WriteString(req.Task + "\n"); err != nil {
		return fmt.Errorf("delegation: write task: %w", err)
	}
	if err := r.stdin.Flush(); err != nil {
		return fmt.Errorf("delegation: flush: %w", err)
	}
	// Close stdin so the sub-agent sees EOF and does not hang waiting
	// for more input. Drop the error-path defer since we are about to
	// start the process, which owns the pipe from here.
	if err := stdinPipe.Close(); err != nil {
		return fmt.Errorf("delegation: close stdin: %w", err)
	}
	return r.cmd.Start()
}

func (r *runner) buildArgs(req *SpawnRequest) []string {
	args := make([]string, len(r.cfg.ArgsTemplate))
	copy(args, r.cfg.ArgsTemplate)
	if req.Model != "" && r.cfg.ModelFlag != "" {
		for i, a := range args {
			if a == r.cfg.ModelFlag {
				if i+1 < len(args) {
					args[i+1] = req.Model
				} else {
					args = append(args, req.Model)
				}
				break
			}
		}
	}
	return args
}

func (r *runner) readOutput() (string, error) {
	var b strings.Builder
	scanner := bufio.NewScanner(r.stdout)
	// Stream-JSON lines can carry large payloads; cap at 16 MiB per line.
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if r.cfg.OutputFormat == FmtStreamJSON {
			b.WriteString(line)
			b.WriteByte('\n')
		} else {
			b.WriteString(line)
		}
	}
	return b.String(), scanner.Err()
}

func (r *runner) wait() error {
	if r.cmd != nil && r.cmd.Process != nil {
		return r.cmd.Wait()
	}
	return nil
}

func (r *runner) kill() {
	if r.cmd != nil && r.cmd.Process != nil {
		_ = r.cmd.Process.Kill()
	}
}

// GatedRunner is the only exported type that can spawn sub-agents.
// Every spawn passes through Engine.Evaluate. Sub-agent output
// containing structured ActionRequests is gated and executed by the
// daemon — sub-agents have zero direct FS/network/terminal access.
type GatedRunner struct {
	cfg    Config
	engine gatekeeper.Gatekeeper
	limit  *Limiter
	sema   *SemaphoreManager
	mu     sync.Mutex
	nextID int
	active map[string]context.CancelFunc
}

// NewGatedRunner creates a gated delegation runner.
func NewGatedRunner(cfg Config, engine gatekeeper.Gatekeeper, limit *Limiter) *GatedRunner {
	return &GatedRunner{
		cfg:    cfg,
		engine: engine,
		limit:  limit,
		active: make(map[string]context.CancelFunc),
	}
}

// SetSemaphoreManager wires the concurrency limiter. The runner uses
// nil-safe semantics: a nil semaphore is a no-op (useful in tests).
func (g *GatedRunner) SetSemaphoreManager(sema *SemaphoreManager) {
	g.sema = sema
}

// Config returns the runner's agent configuration (read-only).
func (g *GatedRunner) Config() Config {
	return g.cfg
}

// Spawn runs a sub-agent task. The spawn is gated through the
// Gatekeeper; if denied, nothing runs.
func (g *GatedRunner) Spawn(ctx context.Context, req *SpawnRequest) (*SpawnResult, error) {
	agentCfg, ok := g.cfg.FindAgent(req.AgentName)
	if !ok {
		return nil, ErrAgentNotFound
	}

	// Gate the spawn through the real Engine (Phase 9).
	ba := blastradius.Action{
		Kind:      "delegation.spawn",
		TargetApp: req.AgentName,
		Body:      req.Task,
	}
	decision, reason := g.engine.Evaluate(ctx, ba)
	if decision != gatekeeper.Allow {
		return nil, fmt.Errorf("%w: %s", ErrGatedDeny, reason)
	}

	// Create a cancellable sub-context so delegate.cancel can interrupt
	// the spawn even when the caller did not supply a deadline.
	spawnCtx, cancel := context.WithCancel(ctx)
	spawnID := g.registerSpawn(cancel)
	defer g.unregisterSpawn(spawnID)

	// Check limits: recursion depth + budget.
	if err := g.limit.CheckSpawn(spawnCtx, req.AgentName, req.Depth, req.Budget); err != nil {
		if errors.Is(err, ErrBudgetExceeded) {
			g.limit.ReleaseBudget(req.AgentName, req.Budget)
		}
		return nil, err
	}

	// Acquire concurrency slot.
	if g.sema != nil {
		if err := g.sema.Acquire(spawnCtx, req.AgentName); err != nil {
			g.limit.ReleaseBudget(req.AgentName, req.Budget)
			return nil, err
		}
		defer g.sema.Release(req.AgentName)
	}

	// Run the sub-agent.
	result, runErr := g.runAgent(spawnCtx, spawnID, agentCfg, req)
	if runErr != nil {
		g.limit.ReleaseBudget(req.AgentName, req.Budget)
	}
	return result, runErr
}

// runAgent executes a single sub-agent run and waits for completion,
// timeout, or cancellation. It returns a non-nil SpawnResult even when
// the sub-agent exits with an error, so callers can inspect Output.
func (g *GatedRunner) runAgent(ctx context.Context, spawnID string, agentCfg AgentConfig, req *SpawnRequest) (*SpawnResult, error) {
	start := time.Now()
	r := newRunner(agentCfg)

	if err := r.start(ctx, req); err != nil {
		return nil, err
	}

	// Wait for completion, timeout, or cancellation.
	done := make(chan readResult)
	go func() {
		out, err := r.readOutput()
		done <- readResult{out: out, err: err}
	}()

	timer := time.NewTimer(agentCfg.Timeout)
	defer timer.Stop()

	var readRes readResult
	select {
	case <-ctx.Done():
		return g.finalizeKilled(r, done, start, spawnID, req, ctx.Err())
	case <-timer.C:
		return g.finalizeKilled(r, done, start, spawnID, req, ErrTimeout)
	case readRes = <-done:
	}

	waitErr := r.wait()
	exitCode := 0
	if r.cmd != nil && r.cmd.ProcessState != nil {
		exitCode = r.cmd.ProcessState.ExitCode()
	}

	result := &SpawnResult{
		AgentName: req.AgentName,
		Task:      req.Task,
		Output:    readRes.out,
		ExitCode:  exitCode,
		Duration:  time.Since(start),
		SpawnID:   spawnID,
	}
	if readRes.err != nil {
		result.Output = fmt.Sprintf("error reading output: %v\n%s", readRes.err, readRes.out)
	}
	if waitErr != nil {
		return result, fmt.Errorf("delegation: sub-agent exited with code %d: %w", exitCode, waitErr)
	}
	return result, nil
}

type readResult struct {
	out string
	err error
}

// finalizeKilled kills a sub-agent, drains the output goroutine, and
// returns a result wrapping the reason (context cancellation or timeout).
func (g *GatedRunner) finalizeKilled(r *runner, done chan readResult, start time.Time, spawnID string, req *SpawnRequest, reason error) (*SpawnResult, error) {
	r.kill()
	// Drain the output goroutine to avoid leaking the reader/pipe.
	_ = r.cmd.Wait()

	var readRes readResult
	select {
	case readRes = <-done:
	case <-time.After(2 * time.Second):
	}
	exitCode := 0
	if r.cmd != nil && r.cmd.ProcessState != nil {
		exitCode = r.cmd.ProcessState.ExitCode()
	}
	return &SpawnResult{
		AgentName: req.AgentName,
		Task:      req.Task,
		Output:    readRes.out,
		ExitCode:  exitCode,
		Duration:  time.Since(start),
		SpawnID:   spawnID,
	}, reason
}

func (g *GatedRunner) registerSpawn(cancel context.CancelFunc) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nextID++
	id := fmt.Sprintf("spawn-%d", g.nextID)
	g.active[id] = cancel
	return id
}

func (g *GatedRunner) unregisterSpawn(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.active, id)
}

// Cancel interrupts a running spawn by SpawnID. Returns true if the spawn
// was known and a cancellation was triggered.
func (g *GatedRunner) Cancel(spawnID string) bool {
	g.mu.Lock()
	cancel, ok := g.active[spawnID]
	g.mu.Unlock()
	if ok && cancel != nil {
		cancel()
		return true
	}
	return false
}

// ActionRequests extracts structured action requests from a result.
// The daemon gates each one before execution.
func (g *GatedRunner) ActionRequests(result *SpawnResult) []ActionRequest {
	if result == nil || result.Output == "" {
		return nil
	}
	var requests []ActionRequest
	// Stream-JSON: each line is a JSON object. Parse for action requests.
	for _, line := range strings.Split(result.Output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var ar ActionRequest
		if err := json.Unmarshal([]byte(line), &ar); err != nil {
			continue
		}
		if ar.Kind != "" {
			ar.AgentName = result.AgentName
			requests = append(requests, ar)
		}
	}
	return requests
}

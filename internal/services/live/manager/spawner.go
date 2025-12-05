package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type processSpawner struct {
	logger logging.ApplicationLogger
}

// NewProcessSpawner creates a new process spawner
func NewProcessSpawner(logger logging.ApplicationLogger) live.ProcessSpawner {
	return &processSpawner{
		logger: logger,
	}
}

// Spawn creates a new kronos run-strategy process
func (ps *processSpawner) Spawn(ctx context.Context, strategy *strategy.Strategy) (*exec.Cmd, error) {
	// Build command: kronos run-strategy --strategy <name>
	// The run-strategy command will look in ./strategies/{strategyName}
	cmd := exec.CommandContext(ctx, "kronos", "run-strategy", "--strategy", strategy.Name)

	// Create new process group (survive parent exit)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Create instance log directory
	instanceLogDir := fmt.Sprintf(".kronos/instances/%s", strategy.Name)
	if err := os.MkdirAll(instanceLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create instance log directory: %w", err)
	}

	// Redirect stdout/stderr to log files (NOT to TUI)
	stdoutLog := fmt.Sprintf("%s/stdout.log", instanceLogDir)
	stderrLog := fmt.Sprintf("%s/stderr.log", instanceLogDir)

	stdoutFile, err := os.OpenFile(stdoutLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout log: %w", err)
	}

	stderrFile, err := os.OpenFile(stderrLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		_ = stdoutFile.Close()
		return nil, fmt.Errorf("failed to open stderr log: %w", err)
	}

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	ps.logger.Info("Spawning strategy process",
		"strategy", strategy.Name,
		"stdout_log", stdoutLog,
		"stderr_log", stderrLog,
	)

	return cmd, nil
}

// AttachMonitor starts monitoring process for crashes
func (ps *processSpawner) AttachMonitor(instance *live.Instance) error {
	if instance.Cmd == nil {
		return fmt.Errorf("command not set on instance")
	}

	ps.logger.Info("Attached monitor to instance", "strategy", instance.StrategyName, "pid", instance.PID)

	return nil
}

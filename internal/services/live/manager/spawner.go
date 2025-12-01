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

	// Set up stdout/stderr (can be redirected to files later)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ps.logger.Info("Spawning strategy process", "strategy", strategy.Name)

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

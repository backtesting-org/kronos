package live

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
)

// InstanceStatus represents the current state of a strategy instance
type InstanceStatus string

const (
	StatusRunning    InstanceStatus = "running"
	StatusStopped    InstanceStatus = "stopped"
	StatusCrashed    InstanceStatus = "crashed"
	StatusRestarting InstanceStatus = "restarting"
)

// Instance represents a running strategy instance
type Instance struct {
	ID              string             `json:"id"`
	StrategyName    string             `json:"strategy_name"`
	StrategyPath    string             `json:"strategy_path"`
	FrameworkRoot   string             `json:"framework_root"`
	PID             int                `json:"pid"`
	Status          InstanceStatus     `json:"status"`
	StartedAt       time.Time          `json:"started_at"`
	LastStatusCheck time.Time          `json:"last_status_check"`
	Restarts        int                `json:"restarts"`
	Error           string             `json:"error"`
	Context         context.Context    `json:"-"`
	Cancel          context.CancelFunc `json:"-"`
	Cmd             *exec.Cmd          `json:"-"`
}

// InstanceManager orchestrates spawning, tracking, and lifecycle of strategy instances
type InstanceManager interface {
	// Start spawns a new strategy instance
	Start(ctx context.Context, strategy *config.Strategy, frameworkRoot string) (*Instance, error)

	// Stop gracefully terminates an instance by ID
	Stop(instanceID string) error

	// StopByStrategyName gracefully terminates an instance by strategy name
	StopByStrategyName(strategyName string) error

	// Kill forcefully terminates an instance
	Kill(instanceID string) error

	// Get retrieves a specific instance
	Get(instanceID string) (*Instance, error)

	// List returns all instances (filtered by status)
	List(status InstanceStatus) ([]*Instance, error)

	// LoadRunning loads instances from state file (after restart)
	LoadRunning(ctx context.Context) error

	// SaveState persists current state to disk
	SaveState() error

	// Shutdown gracefully terminates all instances
	Shutdown(ctx context.Context, timeout time.Duration) error
}

// ProcessSpawner creates and configures child processes with proper isolation
type ProcessSpawner interface {
	// Spawn creates a new kronos run-strategy process
	Spawn(ctx context.Context, strategy *config.Strategy) (*exec.Cmd, error)

	// AttachMonitor starts monitoring process for crashes
	AttachMonitor(instance *Instance) error
}

// StateStore persists and recovers instance state across CLI invocations
type StateStore interface {
	// Load reads persisted state from disk
	Load() ([]*Instance, error)

	// Save writes current state to disk
	Save(instances []*Instance) error

	// GetPath returns the path to the state file
	GetPath() string
}

// SignalHandler handles OS signals and coordinates graceful shutdown
type SignalHandler interface {
	// Handle registers signal handlers
	Handle(ctx context.Context, onSignal func(os.Signal) error) error
}

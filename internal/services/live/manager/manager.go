package manager

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/google/uuid"
)

type instanceManager struct {
	mu          sync.RWMutex
	instances   map[string]*live.Instance
	stateStore  live.StateStore
	spawner     live.ProcessSpawner
	logger      logging.ApplicationLogger
	monitorDone chan struct{}
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(
	stateStore live.StateStore,
	spawner live.ProcessSpawner,
	logger logging.ApplicationLogger,
) live.InstanceManager {
	return &instanceManager{
		instances:   make(map[string]*live.Instance),
		stateStore:  stateStore,
		spawner:     spawner,
		logger:      logger,
		monitorDone: make(chan struct{}),
	}
}

// Start spawns a new strategy instance
func (im *instanceManager) Start(ctx context.Context, strategy *strategy.Strategy, frameworkRoot string) (*live.Instance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Check if already running
	for _, inst := range im.instances {
		if inst.StrategyName == strategy.Name && inst.Status == live.StatusRunning {
			return nil, fmt.Errorf("strategy '%s' already running", strategy.Name)
		}
	}

	// Spawn process
	cmd, err := im.spawner.Spawn(ctx, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to spawn process: %w", err)
	}

	// Create instance
	instCtx, cancel := context.WithCancel(ctx)
	instance := &live.Instance{
		ID:              uuid.New().String(),
		StrategyName:    strategy.Name,
		StrategyPath:    strategy.Path,
		FrameworkRoot:   frameworkRoot,
		Status:          live.StatusRunning,
		StartedAt:       time.Now(),
		LastStatusCheck: time.Now(),
		Context:         instCtx,
		Cancel:          cancel,
		Cmd:             cmd,
	}

	// Start process
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	instance.PID = cmd.Process.Pid

	// Track instance
	im.instances[instance.ID] = instance
	im.logger.Info("Started instance", "strategy", strategy.Name, "id", instance.ID, "pid", instance.PID)

	// Monitor process in background
	go im.monitorProcess(instance)

	// Save state
	_ = im.saveStateLocked()

	return instance, nil
}

// Stop gracefully terminates an instance
func (im *instanceManager) Stop(instanceID string) error {
	im.mu.Lock()
	instance, exists := im.instances[instanceID]
	if !exists {
		im.mu.Unlock()
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	im.mu.Unlock()

	instance.Cancel()

	// Get process handle - either from Cmd (if we spawned it) or by PID (if reattached)
	var process *os.Process
	var err error

	if instance.Cmd != nil && instance.Cmd.Process != nil {
		// We spawned this process - use the Cmd's process handle
		process = instance.Cmd.Process
	} else if instance.PID > 0 {
		// We reattached to this process - find it by PID
		process, err = os.FindProcess(instance.PID)
		if err != nil {
			return fmt.Errorf("failed to find process: %w", err)
		}
	} else {
		return fmt.Errorf("instance has no valid process reference")
	}

	// Send SIGTERM
	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("failed to signal process: %w", err)
	}

	// Wait for graceful exit with timeout
	done := make(chan error, 1)
	go func() {
		if instance.Cmd != nil {
			// If we have Cmd, use Wait() which is cleaner
			done <- instance.Cmd.Wait()
		} else {
			// Otherwise poll for process exit
			for i := 0; i < 100; i++ { // 10 seconds (100 * 100ms)
				time.Sleep(100 * time.Millisecond)
				// Try to signal with signal 0 to check if process exists
				if err := process.Signal(syscall.Signal(0)); err != nil {
					// Process is gone
					done <- nil
					return
				}
			}
			done <- fmt.Errorf("timeout waiting for process to exit")
		}
	}()

	select {
	case <-time.After(10 * time.Second):
		// Force kill if not exited
		im.logger.Warn("Graceful stop timeout, force killing", "instance", instanceID)
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	case <-done:
	}

	im.mu.Lock()
	instance.Status = live.StatusStopped
	instance.PID = 0
	_ = im.saveStateLocked()
	im.mu.Unlock()

	im.logger.Info("Stopped instance", "strategy", instance.StrategyName, "id", instanceID)

	return nil
}

// StopByStrategyName gracefully terminates an instance by strategy name
func (im *instanceManager) StopByStrategyName(strategyName string) error {
	im.mu.RLock()
	var instanceID string

	im.logger.Info("Searching for instance to stop",
		"strategy", strategyName,
		"total_instances", len(im.instances))

	for id, inst := range im.instances {
		im.logger.Debug("Checking instance",
			"id", id,
			"strategy", inst.StrategyName,
			"status", inst.Status)

		if inst.StrategyName == strategyName && inst.Status == live.StatusRunning {
			instanceID = id
			break
		}
	}
	im.mu.RUnlock()

	if instanceID == "" {
		return fmt.Errorf("no running instance found for strategy: %s (instances in memory: %d) - try reloading instances from state",
			strategyName, len(im.instances))
	}

	return im.Stop(instanceID)
}

// Kill forcefully terminates an instance
func (im *instanceManager) Kill(instanceID string) error {
	im.mu.Lock()
	instance, exists := im.instances[instanceID]
	if !exists {
		im.mu.Unlock()
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	im.mu.Unlock()

	instance.Cancel()

	if err := instance.Cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	im.mu.Lock()
	instance.Status = live.StatusStopped
	instance.PID = 0
	_ = im.saveStateLocked()
	im.mu.Unlock()

	im.logger.Info("Killed instance", "strategy", instance.StrategyName, "id", instanceID)

	return nil
}

// Restart stops and restarts an instance
func (im *instanceManager) Restart(instanceID string) error {
	im.mu.RLock()
	instance, exists := im.instances[instanceID]
	if !exists {
		im.mu.RUnlock()
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	strategyName := instance.StrategyName
	im.mu.RUnlock()

	// Stop it
	if err := im.Stop(instanceID); err != nil {
		return err
	}

	// Get strategy
	// TODO: Need to fetch strategy from config
	// For now, we'll rely on the caller to handle this

	im.logger.Info("Restarted instance", "strategy", strategyName, "id", instanceID)

	return nil
}

// Get retrieves a specific instance
func (im *instanceManager) Get(instanceID string) (*live.Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instance, exists := im.instances[instanceID]
	if !exists {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	return instance, nil
}

// List returns all instances (filtered by status)
func (im *instanceManager) List(status live.InstanceStatus) ([]*live.Instance, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var result []*live.Instance

	for _, instance := range im.instances {
		if status == "" || instance.Status == status {
			result = append(result, instance)
		}
	}

	return result, nil
}

// LoadRunning loads instances from state file (after restart)
func (im *instanceManager) LoadRunning(ctx context.Context) error {
	instances, err := im.stateStore.Load()
	if err != nil {
		return err
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	for _, instance := range instances {
		if instance.Status == live.StatusRunning {
			// Verify process still exists
			_, err := os.FindProcess(instance.PID)
			if err != nil {
				instance.Status = live.StatusCrashed
				instance.Error = "Process not found after restart"
				im.logger.Warn("Process not found", "instance", instance.ID, "pid", instance.PID)
				continue
			}

			// Process still alive - reattach monitoring
			instCtx, cancel := context.WithCancel(ctx)
			instance.Context = instCtx
			instance.Cancel = cancel
			im.instances[instance.ID] = instance

			go im.monitorProcess(instance)
			im.logger.Info("Reattached to running instance", "strategy", instance.StrategyName, "pid", instance.PID)
		}
	}

	return im.saveStateLocked()
}

// SaveState persists current state to disk
func (im *instanceManager) SaveState() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	return im.saveStateLocked()
}

// saveStateLocked persists state (must be called with lock held)
func (im *instanceManager) saveStateLocked() error {
	instances := make([]*live.Instance, 0, len(im.instances))
	for _, inst := range im.instances {
		instances = append(instances, inst)
	}

	return im.stateStore.Save(instances)
}

// Shutdown gracefully terminates all instances
func (im *instanceManager) Shutdown(ctx context.Context, timeout time.Duration) error {
	im.mu.RLock()
	instanceIDs := make([]string, 0, len(im.instances))
	for id := range im.instances {
		instanceIDs = append(instanceIDs, id)
	}
	im.mu.RUnlock()

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, len(instanceIDs))

	// Stop all in parallel
	for _, id := range instanceIDs {
		go func(instID string) {
			done <- im.Stop(instID)
		}(id)
	}

	// Wait for all
	var errs []error
	for i := 0; i < len(instanceIDs); i++ {
		select {
		case err := <-done:
			if err != nil {
				errs = append(errs, err)
			}
		case <-shutdownCtx.Done():
			return shutdownCtx.Err()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// monitorProcess monitors a running process for crashes
func (im *instanceManager) monitorProcess(instance *live.Instance) {
	defer im.logger.Info("Monitor stopped", "strategy", instance.StrategyName, "id", instance.ID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-instance.Context.Done():
			return
		case <-ticker.C:
			// Check if process still alive
			if _, err := os.FindProcess(instance.PID); err != nil {
				im.mu.Lock()
				instance.Status = live.StatusCrashed
				instance.Error = "Process exited unexpectedly"
				_ = im.saveStateLocked()
				im.mu.Unlock()

				im.logger.Error("Instance crashed", "strategy", instance.StrategyName, "id", instance.ID)
				return
			}

			im.mu.Lock()
			instance.LastStatusCheck = time.Now()
			im.mu.Unlock()
		}
	}
}

package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/live-trading/pkg/startup"
)

type liveRuntime struct {
	logger       logging.ApplicationLogger
	startup      *startup.Startup
	configLoader config.StartupConfigLoader
}

func NewRuntime(
	logger logging.ApplicationLogger,
	startup *startup.Startup,
	configLoader config.StartupConfigLoader,
) live.Runtime {
	return &liveRuntime{
		logger:       logger,
		startup:      startup,
		configLoader: configLoader,
	}
}

func (r *liveRuntime) Run(strategyDir string) error {
	kronosPath := "kronos.yml"
	cfg, err := r.configLoader.LoadForStrategy(strategyDir, kronosPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	r.logger.Info("Config loaded", "strategy", cfg.Strategy.Name)

	err = r.startup.Start(strategyDir, kronosPath)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	r.logger.Info("SDK startup complete")
	r.logger.Info("Strategy running, keeping process alive...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	r.logger.Info("Received shutdown signal", "signal", sig)

	r.logger.Info("Stopping strategy...")
	if err := r.startup.Stop(); err != nil {
		r.logger.Error("Failed to stop strategy", "error", err)
	}

	r.logger.Info("Shutdown complete")
	return nil
}

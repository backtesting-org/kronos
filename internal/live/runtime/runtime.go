package runtime

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/backtesting-org/kronos-cli/internal/config/connectors"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	sdkRuntime "github.com/backtesting-org/live-trading/pkg/runtime"
)

type liveRuntime struct {
	logger       logging.ApplicationLogger
	runtime      sdkRuntime.Runtime
	connectorSvc connectors.ConnectorService
	strategyConf strategy.StrategyConfig
}

func NewRuntime(
	logger logging.ApplicationLogger,
	runtime sdkRuntime.Runtime,
	connectorSvc connectors.ConnectorService,
	strategyConf strategy.StrategyConfig,
) Runtime {
	return &liveRuntime{
		logger:       logger,
		runtime:      runtime,
		connectorSvc: connectorSvc,
		strategyConf: strategyConf,
	}
}

func (r *liveRuntime) Run(ctx context.Context, strategyDir string) error {
	// 1. Load strategy config from directory
	configPath := filepath.Join(strategyDir, "config.yml")
	strat, err := r.strategyConf.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load strategy config: %w", err)
	}

	r.logger.Info("Loaded strategy config", "name", strat.Name, "exchanges", strat.Exchanges)

	// 2. Get the strategy plugin path
	strategyName := filepath.Base(strategyDir)
	soPath := filepath.Join(strategyDir, strategyName+".so")

	// 3. Get connector configs for all exchanges this strategy needs
	connectorConfigs, err := r.connectorSvc.GetConnectorConfigsForStrategy(strat.Exchanges)
	if err != nil {
		return fmt.Errorf("failed to get connector configs: %w", err)
	}

	r.logger.Info("Prepared connector configs", "count", len(connectorConfigs))

	// 4. Execute the strategy using the SDK runtime
	r.logger.Info("Starting live trading runtime", "strategy", strat.Name, "exchanges", len(connectorConfigs))

	// Run in a goroutine and monitor context for cancellation
	errChan := make(chan error, 1)
	go func() {
		errChan <- r.runtime.Execute(soPath, connectorConfigs)
	}()

	// Wait for either completion or cancellation
	select {
	case <-ctx.Done():
		r.logger.Info("Shutdown signal received, stopping strategy")
		// Context cancelled - return the error immediately
		// The SDK runtime execution in the goroutine will be orphaned and cleaned up by the OS
		return ctx.Err()
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}
	}

	r.logger.Info("Live trading runtime stopped")
	return nil
}

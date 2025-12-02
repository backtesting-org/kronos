package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/config/connectors"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/live-trading/pkg/startup"
)

type liveRuntime struct {
	logger       logging.ApplicationLogger
	startup      startup.Startup
	connectorSvc connectors.ConnectorService
	strategyConf strategy.StrategyConfig
}

func NewRuntime(
	logger logging.ApplicationLogger,
	runtime startup.Startup,
	connectorSvc connectors.ConnectorService,
	strategyConf strategy.StrategyConfig,
) live.Runtime {
	return &liveRuntime{
		logger:       logger,
		startup:      runtime,
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
		return fmt.Errorf("failed to get connector config: %w", err)
	}

	r.logger.Info("Prepared connector configs", "count", len(connectorConfigs))

	// 4. Execute the strategy using the SDK startup
	r.logger.Info("Starting live trading startup", "strategy", strat.Name, "exchanges", len(connectorConfigs))

	assetConfigs := r.convertConfigAssetsToInstruments(strat)

	// Start SDK startup (initializes everything)
	err = r.startup.Start(
		soPath,
		connectorConfigs,
		assetConfigs,
	)
	if err != nil {
		return fmt.Errorf("startup error: %w", err)
	}

	r.logger.Info("✅ SDK startup complete, keeping process alive...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block indefinitely - this child process stays alive trading
	// Strategy execution happens in SDK's background goroutines
	// Only exits when receiving SIGTERM (from manager.Stop()) or manual interrupt
	sig := <-sigChan
	r.logger.Info("Received shutdown signal", "signal", sig)

	return nil
}

// convertConfigAssetsToInstruments converts string instrument names to SDK connector.Instrument enums
func (r *liveRuntime) convertConfigAssetsToInstruments(strat *strategy.Strategy) map[portfolio.Asset][]connector.Instrument {
	instrumentMap := make(map[portfolio.Asset][]connector.Instrument)

	for _, assets := range strat.Assets {
		for _, asset := range assets {
			instruments := make([]connector.Instrument, 0, len(asset.Instruments))

			for _, instStr := range asset.Instruments {
				switch instStr {
				case "spot":
					instruments = append(instruments, connector.TypeSpot)
				case "perpetual":
					instruments = append(instruments, connector.TypePerpetual)
				default:
					r.logger.Warn("Unknown instrument type", "instrument", instStr)
				}
			}

			if len(instruments) > 0 {
				instrumentMap[portfolio.NewAsset(asset.Symbol)] = instruments
			}
		}
	}

	return instrumentMap
}

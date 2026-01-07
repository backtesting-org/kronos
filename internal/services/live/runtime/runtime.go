package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/live-trading/pkg/startup"
)

type liveRuntime struct {
	logger       logging.ApplicationLogger
	startup      startup.Startup
	connectorSvc config.ConnectorService
	strategyConf config.StrategyConfig
}

func NewRuntime(
	logger logging.ApplicationLogger,
	runtime startup.Startup,
	connectorSvc config.ConnectorService,
	strategyConf config.StrategyConfig,
) live.Runtime {
	return &liveRuntime{
		logger:       logger,
		startup:      runtime,
		connectorSvc: connectorSvc,
		strategyConf: strategyConf,
	}
}

func (r *liveRuntime) Run(strategyDir string) error {
	configPath := filepath.Join(strategyDir, "config.yml")
	strat, err := r.strategyConf.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load strategy config: %w", err)
	}

	r.logger.Info("Loaded strategy config", "name", strat.Name, "exchanges", strat.Exchanges)

	strategyName := filepath.Base(strategyDir)
	soPath := filepath.Join(strategyDir, strategyName+".so")

	connectorConfigs, err := r.connectorSvc.GetConnectorConfigsForStrategy(strat.Exchanges)
	if err != nil {
		return fmt.Errorf("failed to get connector config: %w", err)
	}

	r.logger.Info("Prepared connector configs", "count", len(connectorConfigs))
	r.logger.Info("Starting live trading startup", "strategy", strat.Name, "exchanges", len(connectorConfigs))

	assetConfigs := r.convertConfigAssetsToInstruments(strat)

	err = r.startup.Start(soPath, connectorConfigs, assetConfigs)
	if err != nil {
		return fmt.Errorf("startup error: %w", err)
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

func (r *liveRuntime) convertConfigAssetsToInstruments(strat *config.Strategy) map[portfolio.Asset][]connector.Instrument {
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

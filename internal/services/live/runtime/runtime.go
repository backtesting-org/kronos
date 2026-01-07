package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/backtesting-org/kronos-cli/internal/services/monitoring"
	"github.com/backtesting-org/kronos-cli/pkg/live"
	pkgmonitoring "github.com/backtesting-org/kronos-cli/pkg/monitoring"
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
	viewRegistry pkgmonitoring.ViewRegistry
}

func NewRuntime(
	logger logging.ApplicationLogger,
	runtime startup.Startup,
	connectorSvc config.ConnectorService,
	strategyConf config.StrategyConfig,
	viewRegistry pkgmonitoring.ViewRegistry,
) live.Runtime {
	return &liveRuntime{
		logger:       logger,
		startup:      runtime,
		connectorSvc: connectorSvc,
		strategyConf: strategyConf,
		viewRegistry: viewRegistry,
	}
}

func (r *liveRuntime) Run(strategyDir string) error {
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

	r.logger.Info("✅ SDK startup complete")

	// 5. Create shutdown context for graceful termination
	ctx := context.Background()
	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	// 6. Start monitoring server for this instance with shutdown callback
	instanceID := strat.Name // Use strategy name as instance ID
	monitoringServer, err := monitoring.NewServer(
		pkgmonitoring.ServerConfig{
			InstanceID: instanceID,
		},
		r.viewRegistry,
		shutdownCancel, // Pass context cancel function for HTTP shutdown
	)
	if err != nil {
		r.logger.Error("Failed to create monitoring server", "error", err)
		// Continue without monitoring - not fatal
	} else {
		go func() {
			r.logger.Info("Starting monitoring server", "instanceID", instanceID, "socket", monitoringServer.SocketPath())
			if err := monitoringServer.Start(); err != nil {
				r.logger.Error("Monitoring server error", "error", err)
			}
		}()
	}

	r.logger.Info("✅ Strategy running, keeping process alive...")

	// 7. Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 8. Block until shutdown is triggered (either by signal or HTTP /shutdown)
	select {
	case sig := <-sigChan:
		r.logger.Info("Received shutdown signal", "signal", sig)
	case <-shutdownCtx.Done():
		r.logger.Info("Received HTTP shutdown command")
	}

	// 9. Create cleanup context with timeout for shutdown operations
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	// 10. Stop strategy FIRST (graceful SDK shutdown via live-trading)
	r.logger.Info("Stopping strategy...")
	if err := r.startup.Stop(); err != nil {
		r.logger.Error("Failed to stop strategy", "error", err)
		// Continue with cleanup even if this fails
	}

	// 11. Then stop monitoring server
	if monitoringServer != nil {
		r.logger.Info("Stopping monitoring server...")
		if err := monitoringServer.Stop(cleanupCtx); err != nil {
			r.logger.Error("Failed to stop monitoring server", "error", err)
		}
	}

	r.logger.Info("✅ Shutdown complete")
	return nil
}

// convertConfigAssetsToInstruments converts string instrument names to SDK connector.Instrument enums
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

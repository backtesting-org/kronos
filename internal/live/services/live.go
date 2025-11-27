package services

import (
	"context"
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	settings settings.Configuration
	compile  shared.CompileService
	discover shared.StrategyDiscovery
	logger   logging.ApplicationLogger
}

func NewLiveService(
	kronos settings.Configuration,
	compileSvc shared.CompileService,
	discovery shared.StrategyDiscovery,
	logger logging.ApplicationLogger,
) types.LiveService {
	return &liveService{
		settings: kronos,
		compile:  compileSvc,
		discover: discovery,
		logger:   logger,
	}
}

func (s *liveService) FindStrategies() ([]strategy.Strategy, error) {
	strategies, err := s.discover.DiscoverStrategies()

	if err != nil {
		return nil, err
	}

	return strategies, nil
}

func (s *liveService) FindConnectors() []settings.Connector {
	setting, err := s.settings.LoadSettings()

	if err != nil {
		s.logger.Error("Failed to load settings", "error", err)
		return []settings.Connector{}
	}

	if setting == nil {
		return []settings.Connector{}
	}

	return setting.Connectors
}

// ValidateCredentials validates exchange credentials
func (s *liveService) ValidateCredentials(exchangeName string, credentials map[string]string) error {
	// TODO: Add actual validation logic based on exchange type
	return nil
}

// ExecuteStrategy runs the selected strategy with the selected exchange
func (s *liveService) ExecuteStrategy(ctx context.Context, strategy *strategy.Strategy, exchange *settings.Connector) error {
	s.logger.Info("Preparing to execute strategy",
		"strategy", strategy.Name,
		"exchange", exchange.Name,
	)

	// 1. Validate credentials
	if err := s.ValidateCredentials(exchange.Name, exchange.Credentials); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// 2. Connect to exchange
	fmt.Printf("\n🔌 Connecting to %s...\n", exchange.Name)
	// TODO: Initialize connector

	// 3. Load strategy plugin
	pluginPath := fmt.Sprintf("%s/%s.so", strategy.Path, strategy.Name)
	s.logger.Info("Loading strategy plugin", "path", pluginPath)
	// TODO: Load and execute plugin

	// 4. Execute strategy
	fmt.Printf("🚀 Starting strategy: %s\n", strategy.Name)
	fmt.Println("Press Ctrl+C to stop...")

	// Block until context is cancelled
	<-ctx.Done()

	fmt.Println("\n✅ Strategy stopped successfully")
	return nil
}

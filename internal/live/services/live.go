package services

import (
	"context"
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	compile  shared.CompileService
	discover shared.StrategyDiscovery
	config   types.ConfigService
	logger   logging.ApplicationLogger
}

func NewLiveService(
	compileSvc shared.CompileService,
	configSvc types.ConfigService,
	discovery shared.StrategyDiscovery,
	logger logging.ApplicationLogger,
) types.LiveService {
	return &liveService{
		compile:  compileSvc,
		discover: discovery,
		config:   configSvc,
		logger:   logger,
	}
}

// DiscoverStrategies finds and compiles all available strategies
func (s *liveService) DiscoverStrategies() ([]types.Strategy, error) {
	s.logger.Info("Discovering strategies...")
	return s.discover.DiscoverStrategies()
}

// LoadConnectors loads exchange configurations from kronos.yml
func (s *liveService) LoadConnectors() (types.Connectors, error) {
	s.logger.Info("Loading exchange configuration...")
	connectors, err := s.config.LoadExchangeCredentials()
	if err != nil {
		return types.Connectors{}, err
	}
	return connectors, nil
}

// ValidateCredentials validates exchange credentials
func (s *liveService) ValidateCredentials(exchangeName string, credentials map[string]string) error {
	// TODO: Add actual validation logic based on exchange type
	return nil
}

// ExecuteStrategy runs the selected strategy with the selected exchange
func (s *liveService) ExecuteStrategy(ctx context.Context, strategy *types.Strategy, exchange *types.ExchangeConfig) error {
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

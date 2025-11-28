package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/backtesting-org/kronos-cli/internal/config/connectors"
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	settings         settings.Configuration
	connectorService connectors.ConnectorService
	compile          shared.CompileService
	discover         shared.StrategyDiscovery
	logger           logging.ApplicationLogger
}

func NewLiveService(
	kronos settings.Configuration,
	connectorService connectors.ConnectorService,
	compileSvc shared.CompileService,
	discovery shared.StrategyDiscovery,
	logger logging.ApplicationLogger,
) types.LiveService {
	return &liveService{
		settings:         kronos,
		connectorService: connectorService,
		compile:          compileSvc,
		discover:         discovery,
		logger:           logger,
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

// ValidateStrategy checks if the strategy can be executed (has valid connectors)
func (s *liveService) ValidateStrategy(strat *strategy.Strategy) error {
	_, err := s.connectorService.GetConnectorConfigsForStrategy(strat.Exchanges)
	if err != nil {
		return fmt.Errorf("Missing connectors: %v\n\nPlease ensure these exchanges are:\n  • Configured in exchanges.yml\n  • Marked as enabled\n  • Available in the SDK", strat.Exchanges)
	}
	return nil
}

// ExecuteStrategy runs the selected strategy with all its configured exchanges
func (s *liveService) ExecuteStrategy(ctx context.Context, strat *strategy.Strategy, connector *settings.Connector) error {
	// 1. Pre-validate that we have connectors for this strategy's exchanges
	connectorConfigs, err := s.connectorService.GetConnectorConfigsForStrategy(strat.Exchanges)
	if err != nil {
		return fmt.Errorf("cannot start strategy '%s': %w\n\nPlease check:\n- exchanges.yml has entries for: %v\n- Required exchanges are enabled\n- Exchange connectors are available in the SDK",
			strat.Name, err, strat.Exchanges)
	}

	s.logger.Info("Validated connector configs", "strategy", strat.Name, "connectors", len(connectorConfigs))

	// 2. Compile strategy if needed
	if err := s.compile.CompileStrategy(strat.Path); err != nil {
		return fmt.Errorf("failed to compile strategy: %w", err)
	}

	// 3. Spawn kronos with run-strategy subcommand
	args := []string{
		"run-strategy",
		"--strategy", strat.Name,
	}

	if strat.Execution.DryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command("kronos", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	s.logger.Info("Spawning strategy instance",
		"strategy", strat.Name,
		"exchanges", strat.Exchanges,
	)

	// Start in background - don't wait for it to finish
	return cmd.Start()
}

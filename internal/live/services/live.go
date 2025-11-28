package services

import (
	"context"
	"errors"
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
		// Check if it's a StrategyValidationError so we can provide detailed feedback
		var sve *connectors.StrategyValidationError
		if errors.As(err, &sve) {
			// Build a detailed error message from the specific problems
			msg := fmt.Sprintf("Cannot start '%s' - missing or invalid connectors:\n\n", strat.Name)

			notFound := sve.GetExchangesByProblem("not_found")
			notEnabled := sve.GetExchangesByProblem("not_enabled")
			missingCreds := sve.GetExchangesByProblem("missing_credentials")
			invalidConfig := sve.GetExchangesByProblem("invalid_config")

			if len(notFound) > 0 {
				msg += fmt.Sprintf("❌ Not in SDK: %v\n   (Exchange connector not available)\n\n", notFound)
			}
			if len(notEnabled) > 0 {
				msg += fmt.Sprintf("❌ Not enabled: %v\n   (Add to exchanges.yml and set enabled: true)\n\n", notEnabled)
			}
			if len(missingCreds) > 0 {
				msg += fmt.Sprintf("❌ Missing credentials: %v\n", missingCreds)
				for _, ex := range missingCreds {
					if valErr := sve.GetExchangeError(ex); valErr != nil && len(valErr.Missing) > 0 {
						msg += fmt.Sprintf("   • %s needs: %v\n", ex, valErr.Missing)
					}
				}
				msg += "\n"
			}
			if len(invalidConfig) > 0 {
				msg += fmt.Sprintf("❌ Invalid config: %v\n", invalidConfig)
				for _, ex := range invalidConfig {
					if valErr := sve.GetExchangeError(ex); valErr != nil {
						if valErr.SDKValidationErr != "" {
							msg += fmt.Sprintf("   • %s: %s\n", ex, valErr.SDKValidationErr)
						}
						for field, reason := range valErr.InvalidFields {
							msg += fmt.Sprintf("   • %s.%s: %s\n", ex, field, reason)
						}
					}
				}
			}

			return fmt.Errorf(msg)
		}

		return fmt.Errorf("failed to validate connectors: %w", err)
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

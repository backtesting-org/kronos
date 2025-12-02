package live

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/backtesting-org/kronos-cli/internal/config/connectors"
	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	strategyTypes "github.com/backtesting-org/kronos-cli/pkg/strategy"

	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-cli/pkg/live"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type LiveService interface {
	FindStrategies() ([]strategy.Strategy, error)
	FindConnectors() []settings.Connector
	ValidateStrategy(strat *strategy.Strategy) error
	ExecuteStrategy(ctx context.Context, strategy *strategy.Strategy, exchange *settings.Connector) error
}

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	settings         settings.Configuration
	connectorService connectors.ConnectorService
	compile          strategyTypes.CompileService
	discover         shared.StrategyDiscovery
	logger           logging.ApplicationLogger
	manager          live.InstanceManager
}

func NewLiveService(
	kronos settings.Configuration,
	connectorService connectors.ConnectorService,
	compileSvc strategyTypes.CompileService,
	discovery shared.StrategyDiscovery,
	logger logging.ApplicationLogger,
	manager live.InstanceManager,
) LiveService {
	return &liveService{
		settings:         kronos,
		connectorService: connectorService,
		compile:          compileSvc,
		discover:         discovery,
		logger:           logger,
		manager:          manager,
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

	// 3. Get current working directory as framework root
	frameworkRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// 4. Start instance using manager (replaces cmd.Start())
	s.logger.Info("Starting strategy instance via manager",
		"strategy", strat.Name,
		"exchanges", strat.Exchanges,
		"framework_root", frameworkRoot,
	)

	_, err = s.manager.Start(ctx, strat, frameworkRoot)
	return err
}

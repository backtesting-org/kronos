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

// ExecuteStrategy runs the selected strategy with all its configured exchanges
func (s *liveService) ExecuteStrategy(ctx context.Context, strat *strategy.Strategy, connector *settings.Connector) error {
	// 1. Compile strategy if needed
	if err := s.compile.CompileStrategy(strat.Path); err != nil {
		return fmt.Errorf("failed to compile strategy: %w", err)
	}

	// 2. Build a new kronos-live binary for this strategy instance
	kronosLivePath := fmt.Sprintf("%s/kronos-live-%s", strat.Path, strat.Name)

	s.logger.Info("Building kronos-live binary for strategy", "output", kronosLivePath)

	buildCmd := exec.Command("go", "build",
		"-o", kronosLivePath,
		"./cmd/kronos-live",
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build kronos-live binary: %w", err)
	}

	// 3. Build command - just pass the strategy directory
	args := []string{
		"run",
		"--strategy-dir", strat.Path,
	}

	if strat.Execution.DryRun {
		args = append(args, "--dry-run")
	}

	// 4. Execute the new binary instance
	cmd := exec.Command(kronosLivePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	s.logger.Info("Starting live trading instance",
		"strategy", strat.Name,
		"exchanges", strat.Exchanges,
		"binary", kronosLivePath,
	)

	return cmd.Run()
}

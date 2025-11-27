package shared

import (
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type StrategyDiscovery interface {
	DiscoverStrategies() ([]types.Strategy, error)
}

type strategyDiscovery struct {
	compileSvc CompileService
	logger     logging.ApplicationLogger
}

func NewStrategyDiscovery(
	compileSvc CompileService,
	logger logging.ApplicationLogger,
) StrategyDiscovery {
	return &strategyDiscovery{
		compileSvc: compileSvc,
		logger:     logger,
	}
}

// DiscoverStrategies finds and compiles strategies, returns them with compilation status
func (s *strategyDiscovery) DiscoverStrategies() ([]types.Strategy, error) {
	// Pre-compile all strategies
	s.logger.Info("Discovering strategies")
	compileErrors := s.compileSvc.PreCompileStrategies("./strategies")

	// Discover strategies
	strategies, err := types.DiscoverStrategies()
	if err != nil {
		return []types.Strategy{}, nil // Return empty list, not error
	}

	// Apply compilation errors to strategies
	for i := range strategies {
		if compErr, hasError := compileErrors[strategies[i].Name]; hasError {
			strategies[i].Status = types.StatusError
			strategies[i].Error = compErr.Error()
		}
	}

	return strategies, nil
}

package shared

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/types"
)

type StrategyDiscovery interface {
	DiscoverStrategies() ([]types.Strategy, error)
}

type strategyDiscovery struct {
	compileSvc CompileService
}

func NewStrategyDiscovery(compileSvc CompileService) StrategyDiscovery {
	return &strategyDiscovery{
		compileSvc: compileSvc,
	}
}

// DiscoverStrategies finds and compiles strategies, returns them with compilation status
func (s *strategyDiscovery) DiscoverStrategies() ([]types.Strategy, error) {
	// Pre-compile all strategies
	fmt.Println("🔍 Checking strategies...")
	compileErrors := s.compileSvc.PreCompileStrategies("./strategies")
	fmt.Println()

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

package shared

import (
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type StrategyDiscovery interface {
	DiscoverStrategies() ([]strategy.Strategy, error)
}

type strategyDiscovery struct {
	strategyConfig strategy.StrategyConfig
	logger         logging.ApplicationLogger
}

func NewStrategyDiscovery(
	strategyConfig strategy.StrategyConfig,
	logger logging.ApplicationLogger,
) StrategyDiscovery {
	return &strategyDiscovery{
		strategyConfig: strategyConfig,
		logger:         logger,
	}
}

// DiscoverStrategies finds and compiles strategies
func (s *strategyDiscovery) DiscoverStrategies() ([]strategy.Strategy, error) {
	strategies, err := s.strategyConfig.FindStrategies()

	if err != nil {
		s.logger.Error("Failed to discover strategies", "error", err)
		return nil, err
	}

	return strategies, nil
}

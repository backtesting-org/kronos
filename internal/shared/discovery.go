package shared

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/config"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type StrategyDiscovery interface {
	DiscoverStrategies() ([]config.Strategy, error)
}

type strategyDiscovery struct {
	strategyConfig config.StrategyConfig
	logger         logging.ApplicationLogger
}

func NewStrategyDiscovery(
	strategyConfig config.StrategyConfig,
	logger logging.ApplicationLogger,
) StrategyDiscovery {
	return &strategyDiscovery{
		strategyConfig: strategyConfig,
		logger:         logger,
	}
}

// DiscoverStrategies finds and compiles strategies
func (s *strategyDiscovery) DiscoverStrategies() ([]config.Strategy, error) {
	strategies, err := s.strategyConfig.FindStrategies()

	if err != nil {
		s.logger.Error("Failed to discover strategies", "error", err)
		return nil, err
	}

	return strategies, nil
}

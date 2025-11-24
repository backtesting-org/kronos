package services

import (
	"github.com/backtesting-org/kronos-cli/internal/live"
)

// LiveService handles live trading operations
type LiveService struct{}

func NewLiveService() *LiveService {
	return &LiveService{}
}

func (s *LiveService) RunSelectionTUI() error {
	return live.RunSelectionTUI()
}

func (s *LiveService) GetStrategies() ([]live.Strategy, error) {
	// Try to discover strategies from ./strategies directory
	strategies, err := live.DiscoverStrategies()
	if err != nil {
		// Fall back to mock strategies if discovery fails
		return live.GetMockStrategies(), nil
	}
	return strategies, nil
}

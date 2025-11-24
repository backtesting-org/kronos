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
	// For now, returns mock data
	return live.GetMockStrategies(), nil
}

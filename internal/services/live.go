package services

import (
	"github.com/backtesting-org/kronos-cli/internal/live"
)

// LiveService handles live trading operations
type LiveService struct{
	compileSvc *CompileService
}

func NewLiveService(compileSvc *CompileService) *LiveService {
	return &LiveService{
		compileSvc: compileSvc,
	}
}

func (s *LiveService) RunSelectionTUI() error {
	return live.RunSelectionTUI(s.compileSvc)
}

func (s *LiveService) GetStrategies() ([]live.Strategy, error) {
	// Discover strategies from ./strategies directory
	return live.DiscoverStrategies()
}

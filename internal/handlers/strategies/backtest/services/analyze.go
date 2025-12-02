package services

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
)

// analyzeService handles result analysis
type analyzeService struct{}

func NewAnalyzeService() types.AnalyzeService {
	return &analyzeService{}
}

func (s *analyzeService) AnalyzeResults(path string) error {
	// TODO: Implement result analysis
	return nil
}

package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/handlers/strategies/backtest/types"
	"github.com/spf13/cobra"
)

// analyzeHandler handles the analyze command
type analyzeHandler struct {
	analyzeService types.AnalyzeService
}

func NewAnalyzeHandler(analyzeService types.AnalyzeService) types.AnalyzeHandler {
	return &analyzeHandler{
		analyzeService: analyzeService,
	}
}

func (h *analyzeHandler) Handle(cmd *cobra.Command, args []string) error {
	resultsPath, _ := cmd.Flags().GetString("path")
	if resultsPath == "" {
		resultsPath = "./results"
	}

	return h.analyzeService.AnalyzeResults(resultsPath)
}

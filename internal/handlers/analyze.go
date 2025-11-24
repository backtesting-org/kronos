package handlers

import (
	"github.com/backtesting-org/kronos-cli/internal/services"
	"github.com/spf13/cobra"
)

// AnalyzeHandler handles the analyze command
type AnalyzeHandler struct {
	analyzeService *services.AnalyzeService
}

func NewAnalyzeHandler(analyzeService *services.AnalyzeService) *AnalyzeHandler {
	return &AnalyzeHandler{
		analyzeService: analyzeService,
	}
}

func (h *AnalyzeHandler) Handle(cmd *cobra.Command, args []string) error {
	resultsPath, _ := cmd.Flags().GetString("path")
	if resultsPath == "" {
		resultsPath = "./results"
	}

	return h.analyzeService.AnalyzeResults(resultsPath)
}

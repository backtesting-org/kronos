package types

import "github.com/spf13/cobra"

type AnalyzeHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

type BacktestHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

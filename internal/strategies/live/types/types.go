package types

import (
	"context"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-cli/internal/config/strategy"
	"github.com/spf13/cobra"
)

type LiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

type LiveService interface {
	FindStrategies() ([]strategy.Strategy, error)
	FindConnectors() []settings.Connector
	ValidateStrategy(strat *strategy.Strategy) error
	ExecuteStrategy(ctx context.Context, strategy *strategy.Strategy, exchange *settings.Connector) error
}

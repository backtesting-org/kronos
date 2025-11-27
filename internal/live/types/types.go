package types

import (
	"context"

	"github.com/spf13/cobra"
)

type LiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

type ConfigService interface {
	LoadExchangeCredentials() (Connectors, error)
}

type LiveService interface {
	DiscoverStrategies() ([]Strategy, error)
	LoadConnectors() (Connectors, error)
	ValidateCredentials(exchangeName string, credentials map[string]string) error
	ExecuteStrategy(ctx context.Context, strategy *Strategy, exchange *ExchangeConfig) error
}

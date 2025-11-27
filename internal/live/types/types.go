package types

import (
	"github.com/spf13/cobra"
)

type LiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

type ConfigService interface {
	LoadExchangeCredentials() (Connectors, error)
}

type LiveService interface {
	RunSelectionTUI() error
	GetStrategies() ([]Strategy, error)
}

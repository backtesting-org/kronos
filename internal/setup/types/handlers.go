package types

import "github.com/spf13/cobra"

type InitHandler interface {
	Handle(cmd *cobra.Command, args []string) error
	HandleWithStrategy(strategyExample, name string) error
}

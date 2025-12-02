package live

import (
	"github.com/spf13/cobra"
)

type LiveHandler interface {
	Handle(cmd *cobra.Command, args []string) error
}

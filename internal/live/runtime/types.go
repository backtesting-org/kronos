package runtime

import "context"

// Runtime is the interface for the live trading runtime
type Runtime interface {
	Run(ctx context.Context, strategyDir string) error
}

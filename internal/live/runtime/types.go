package runtime

import "context"

// Runtime is the interface for the live trading startup
type Runtime interface {
	Run(ctx context.Context, strategyDir string) error
}

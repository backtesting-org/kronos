package monitoring

import (
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// ViewRegistry aggregates runtime data from SDK stores and exposes it for monitoring.
// This interface is implemented in the SDK and used by the monitoring server.
type ViewRegistry interface {
	GetPnLView() *PnLView
	GetPositionsView() *strategy.StrategyExecution
	GetOrderbookView(symbol string) *connector.OrderBook
	GetRecentTrades(limit int) []connector.Trade
	GetMetrics() *StrategyMetrics
	GetHealth() *health.SystemHealthReport
	GetAvailableAssets() []AssetExchange
}

// AssetExchange represents an asset on a specific exchange
type AssetExchange struct {
	Asset    string `json:"asset"`
	Exchange string `json:"exchange"`
}

// ViewQuerier queries views from running strategy instances via Unix socket.
// This interface is implemented in the CLI to query remote strategy processes.
type ViewQuerier interface {
	// QueryPnL retrieves PnL snapshot from a running instance
	QueryPnL(instanceID string) (*PnLView, error)

	// QueryPositions retrieves active positions from a running instance
	QueryPositions(instanceID string) (*strategy.StrategyExecution, error)

	// QueryOrderbook retrieves orderbook for an asset/exchange from a running instance
	QueryOrderbook(instanceID, asset, exchange string) (*connector.OrderBook, error)

	// QueryRecentTrades retrieves recent trades from a running instance
	QueryRecentTrades(instanceID string, limit int) ([]connector.Trade, error)

	// QueryMetrics retrieves strategy metrics from a running instance
	QueryMetrics(instanceID string) (*StrategyMetrics, error)

	// QueryAvailableAssets retrieves the list of assets being traded
	QueryAvailableAssets(instanceID string) ([]AssetExchange, error)

	// HealthCheck verifies instance is responsive
	HealthCheck(instanceID string) error

	// Shutdown sends shutdown command to instance (graceful HTTP-based shutdown)
	Shutdown(instanceID string) error

	// ListInstances returns all instance IDs that have active sockets
	ListInstances() ([]string, error)
}

package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	monitoring2 "github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/monitoring/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

// querier implements ViewQuerier - queries running strategy instances via Unix socket
type querier struct {
	socketDir string
	timeout   time.Duration
}

// NewQuerier creates a new ViewQuerier client
func NewQuerier() monitoring2.ViewQuerier {
	homeDir, _ := os.UserHomeDir()
	return &querier{
		socketDir: filepath.Join(homeDir, ".kronos", "sockets"),
		timeout:   5 * time.Second,
	}
}

// NewQuerierWithConfig creates a ViewQuerier with custom socket directory
func NewQuerierWithConfig(socketDir string, timeout time.Duration) monitoring2.ViewQuerier {
	return &querier{
		socketDir: socketDir,
		timeout:   timeout,
	}
}

// getClient creates an HTTP client that connects via Unix socket
func (q *querier) getClient(instanceID string) (*http.Client, string, error) {
	socketPath := filepath.Join(q.socketDir, fmt.Sprintf("%s.sock", instanceID))

	// Check socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("instance %s not found (socket does not exist)", instanceID)
	}

	client := &http.Client{
		Timeout: q.timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return client, socketPath, nil
}

// doRequest performs an HTTP request to the instance
func (q *querier) doRequest(instanceID, path string, result interface{}) error {
	client, _, err := q.getClient(instanceID)
	if err != nil {
		return err
	}

	// Use "http://unix" as dummy host - actual connection is via socket
	resp, err := client.Get(fmt.Sprintf("http://unix%s", path))
	if err != nil {
		return fmt.Errorf("failed to connect to instance %s: %w", instanceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("instance %s returned status %d", instanceID, resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// QueryPnL retrieves PnL snapshot from a running instance
func (q *querier) QueryPnL(instanceID string) (*monitoring2.PnLView, error) {
	var result monitoring2.PnLView
	if err := q.doRequest(instanceID, "/api/pnl", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryPositions retrieves active positions from a running instance
func (q *querier) QueryPositions(instanceID string) (*strategy.StrategyExecution, error) {
	var result strategy.StrategyExecution
	if err := q.doRequest(instanceID, "/api/positions", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryOrderbook retrieves orderbook for an asset/exchange from a running instance
func (q *querier) QueryOrderbook(instanceID, asset, exchange string) (*connector.OrderBook, error) {
	var result connector.OrderBook
	path := fmt.Sprintf("/api/orderbook?asset=%s&exchange=%s", asset, exchange)
	if err := q.doRequest(instanceID, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryAvailableAssets retrieves the list of assets being traded
func (q *querier) QueryAvailableAssets(instanceID string) ([]monitoring2.AssetExchange, error) {
	var result []monitoring2.AssetExchange
	if err := q.doRequest(instanceID, "/api/assets", &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryRecentTrades retrieves recent trades from a running instance
func (q *querier) QueryRecentTrades(instanceID string, limit int) ([]connector.Trade, error) {
	var result []connector.Trade
	path := fmt.Sprintf("/api/trades?limit=%d", limit)
	if err := q.doRequest(instanceID, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryMetrics retrieves strategy metrics from a running instance
func (q *querier) QueryMetrics(instanceID string) (*monitoring2.StrategyMetrics, error) {
	var result monitoring2.StrategyMetrics
	if err := q.doRequest(instanceID, "/api/metrics", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryHealth retrieves health status from a running instance
func (q *querier) QueryHealth(instanceID string) (*health.SystemHealthReport, error) {
	var result health.SystemHealthReport
	if err := q.doRequest(instanceID, "/health", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// HealthCheck verifies instance is responsive
func (q *querier) HealthCheck(instanceID string) error {
	_, err := q.QueryHealth(instanceID)
	return err
}

// Shutdown sends graceful shutdown command to instance via HTTP
// Note: This sends the command to the strategy process's monitoring server
func (q *querier) Shutdown(instanceID string) error {
	client, _, err := q.getClient(instanceID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://unix/shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to create shutdown request: %w", err)
	}

	client.Timeout = 2 * time.Second

	resp, err := client.Do(req)
	if err != nil {
		// If connection fails, the process might have already shut down
		// This is actually success for our purposes
		return nil
	}
	defer resp.Body.Close()

	return nil
}

// ListInstances returns all instance IDs that have active sockets
func (q *querier) ListInstances() ([]string, error) {
	entries, err := os.ReadDir(q.socketDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read socket directory: %w", err)
	}

	var instances []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sock" {
			instanceID := entry.Name()[:len(entry.Name())-5] // Remove .sock extension
			instances = append(instances, instanceID)
		}
	}

	return instances, nil
}

// QueryProfilingStats retrieves profiling statistics from a running instance
func (q *querier) QueryProfilingStats(instanceID string) (*monitoring2.ProfilingStats, error) {
	var result monitoring2.ProfilingStats
	if err := q.doRequest(instanceID, "/profiling/stats", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QueryRecentExecutions retrieves recent strategy executions with timing data
func (q *querier) QueryRecentExecutions(instanceID string, limit int) ([]monitoring2.ProfilingMetrics, error) {
	var result []monitoring2.ProfilingMetrics
	path := fmt.Sprintf("/profiling/executions?limit=%d", limit)
	if err := q.doRequest(instanceID, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

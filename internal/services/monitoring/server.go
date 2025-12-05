package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
)

// server implements the Server interface
type server struct {
	config       monitoring.ServerConfig
	viewRegistry monitoring.ViewRegistry
	socketPath   string
	listener     net.Listener
	httpServer   *http.Server
	mu           sync.Mutex
	started      bool
}

// NewServer creates a new monitoring server
func NewServer(config monitoring.ServerConfig, viewRegistry monitoring.ViewRegistry) (monitoring.Server, error) {
	if config.InstanceID == "" {
		return nil, fmt.Errorf("instance ID is required")
	}

	socketDir := config.SocketDir
	if socketDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		socketDir = filepath.Join(homeDir, ".kronos", "sockets")
	}

	// Ensure socket directory exists
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	socketPath := filepath.Join(socketDir, fmt.Sprintf("%s.sock", config.InstanceID))

	return &server{
		config:       config,
		viewRegistry: viewRegistry,
		socketPath:   socketPath,
	}, nil
}

// Start begins listening on the Unix socket
func (s *server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}

	// Remove old socket if exists
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		s.mu.Unlock()
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create unix socket: %w", err)
	}

	// Set socket permissions (owner-only access)
	if err := os.Chmod(s.socketPath, 0600); err != nil {
		_ = listener.Close()
		s.mu.Unlock()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.listener = listener
	s.started = true

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/pnl", s.handlePnL)
	mux.HandleFunc("/api/positions", s.handlePositions)
	mux.HandleFunc("/api/orderbook", s.handleOrderbook)
	mux.HandleFunc("/api/trades", s.handleTrades)
	mux.HandleFunc("/api/metrics", s.handleMetrics)
	mux.HandleFunc("/api/assets", s.handleAssets)

	s.httpServer = &http.Server{Handler: mux}
	s.mu.Unlock()

	// Serve blocks until the server is stopped
	if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	var errs []error

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown http server: %w", err))
		}
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close listener: %w", err))
		}
	}

	// Remove socket file
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("failed to remove socket: %w", err))
	}

	s.started = false

	if len(errs) > 0 {
		return fmt.Errorf("stop errors: %v", errs)
	}

	return nil
}

// SocketPath returns the path to the Unix socket
func (s *server) SocketPath() string {
	return s.socketPath
}

// HTTP Handlers

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	healthReport := s.viewRegistry.GetHealth()
	if healthReport == nil {
		healthReport = &health.SystemHealthReport{OverallState: health.StateConnected}
	}

	s.writeJSON(w, healthReport)
}

func (s *server) handlePnL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pnl := s.viewRegistry.GetPnLView()
	if pnl == nil {
		s.writeJSON(w, struct{}{})
		return
	}

	s.writeJSON(w, pnl)
}

func (s *server) handlePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	positions := s.viewRegistry.GetPositionsView()
	if positions == nil {
		s.writeJSON(w, struct{}{})
		return
	}

	s.writeJSON(w, positions)
}

func (s *server) handleOrderbook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	if asset == "" {
		http.Error(w, "asset parameter required", http.StatusBadRequest)
		return
	}

	orderbook := s.viewRegistry.GetOrderbookView(asset)
	if orderbook == nil {
		http.Error(w, "orderbook not found", http.StatusNotFound)
		return
	}

	s.writeJSON(w, orderbook)
}

func (s *server) handleTrades(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	trades := s.viewRegistry.GetRecentTrades(limit)
	if trades == nil {
		trades = []connector.Trade{}
	}

	s.writeJSON(w, trades)
}

func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.viewRegistry.GetMetrics()
	if metrics == nil {
		s.writeJSON(w, &monitoring.StrategyMetrics{})
		return
	}

	s.writeJSON(w, metrics)
}

func (s *server) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	assets := s.viewRegistry.GetAvailableAssets()
	if assets == nil {
		assets = []monitoring.AssetExchange{}
	}

	s.writeJSON(w, assets)
}

func (s *server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

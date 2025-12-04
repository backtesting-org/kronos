package monitoring

import "context"

// Server exposes strategy runtime views via Unix domain socket HTTP server.
// It allows the CLI to query running strategy instances for real-time data.
type Server interface {
	// Start begins listening on the Unix socket. Blocks until stopped.
	Start() error

	// Stop gracefully shuts down the server
	Stop(ctx context.Context) error

	// SocketPath returns the path to the Unix socket
	SocketPath() string
}

// ServerConfig holds configuration for the monitoring server
type ServerConfig struct {
	// InstanceID is the unique identifier for this strategy instance
	InstanceID string

	// SocketDir is the directory where socket files are created
	// Defaults to ~/.kronos/sockets/
	SocketDir string
}

package services

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/backtesting-org/kronos-cli/internal/live/handlers"
	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/backtesting-org/kronos-cli/internal/shared"
	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
	tea "github.com/charmbracelet/bubbletea"
)

type LiveService interface {
	RunSelectionTUI() error
	DiscoverStrategies() ([]types.Strategy, error)
	LoadConnectors() (types.Connectors, error)
	ValidateCredentials(exchangeName string, credentials map[string]string) error
	ExecuteStrategy(ctx context.Context, strategy *types.Strategy, exchange *types.ExchangeConfig) error
}

// liveService orchestrates live trading by coordinating other services
type liveService struct {
	compileSvc shared.CompileService
	configSvc  types.ConfigService
	logger     logging.ApplicationLogger
}

func NewLiveService(
	compileSvc shared.CompileService,
	configSvc types.ConfigService,
	logger logging.ApplicationLogger,
) LiveService {
	return &liveService{
		compileSvc: compileSvc,
		configSvc:  configSvc,
		logger:     logger,
	}
}

// DiscoverStrategies finds and compiles strategies, returns them with compilation status
func (s *liveService) DiscoverStrategies() ([]types.Strategy, error) {
	// Pre-compile all strategies
	fmt.Println("🔍 Checking strategies...")
	compileErrors := s.compileSvc.PreCompileStrategies("./strategies")
	fmt.Println()

	// Discover strategies
	strategies, err := types.DiscoverStrategies()
	if err != nil {
		return []types.Strategy{}, nil // Return empty list, not error
	}

	// Apply compilation errors to strategies
	for i := range strategies {
		if compErr, hasError := compileErrors[strategies[i].Name]; hasError {
			strategies[i].Status = types.StatusError
			strategies[i].Error = compErr.Error()
		}
	}

	return strategies, nil
}

// LoadConnectors loads exchange configurations from kronos.yml
func (s *liveService) LoadConnectors() (types.Connectors, error) {
	s.logger.Info("Loading exchange configuration...")

	connectors, err := s.configSvc.LoadExchangeCredentials()
	if err != nil {
		return types.Connectors{}, fmt.Errorf("failed to load connectors: %w", err)
	}

	return connectors, nil
}

// ValidateCredentials validates exchange credentials
func (s *liveService) ValidateCredentials(exchangeName string, credentials map[string]string) error {
	// TODO: Add actual validation logic based on exchange type
	return nil
}

// RunSelectionTUI orchestrates the TUI selection flow
func (s *liveService) RunSelectionTUI() error {
	// 1. Load connectors configuration
	connectors, err := s.LoadConnectors()
	if err != nil {
		return err
	}

	// 2. Discover and compile strategies
	strategies, err := s.DiscoverStrategies()
	if err != nil {
		return err
	}

	// 3. Show TUI for strategy selection (pure view layer)
	m := handlers.NewSelectionModel(strategies, connectors)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// 4. Handle selection result
	if model, ok := finalModel.(handlers.SelectionModel); ok {
		if model.Err() != nil {
			if model.Err().Error() == "INIT_PROJECT_REQUESTED" {
				return fmt.Errorf("INIT_PROJECT_REQUESTED")
			}
			return model.Err()
		}

		// User selected strategy + exchange
		if model.Selected() != nil && model.SelectedExchange() != nil {
			// Set up context for execution
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle Ctrl+C gracefully
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\n\n🛑 Stopping strategy...")
				cancel()
			}()

			// Execute strategy with selected exchange
			return s.ExecuteStrategy(ctx, model.Selected(), model.SelectedExchange())
		}
	}

	return nil
}

// ExecuteStrategy runs the selected strategy with the selected exchange
func (s *liveService) ExecuteStrategy(ctx context.Context, strategy *types.Strategy, exchange *types.ExchangeConfig) error {
	s.logger.Info("Preparing to execute strategy",
		"strategy", strategy.Name,
		"exchange", exchange.Name,
	)

	// 1. Validate credentials
	if err := s.ValidateCredentials(exchange.Name, exchange.Credentials); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// 2. Connect to exchange
	fmt.Printf("\n🔌 Connecting to %s...\n", exchange.Name)
	// TODO: Initialize connector

	// 3. Load strategy plugin
	pluginPath := fmt.Sprintf("%s/%s.so", strategy.Path, strategy.Name)
	s.logger.Info("Loading strategy plugin", "path", pluginPath)
	// TODO: Load and execute plugin

	// 4. Execute strategy
	fmt.Printf("🚀 Starting strategy: %s\n", strategy.Name)
	fmt.Println("Press Ctrl+C to stop...")

	// Block until context is cancelled
	<-ctx.Done()

	fmt.Println("\n✅ Strategy stopped successfully")
	return nil
}

//
//	// Auto-compile strategy if needed
//	if compileSvc != nil {
//		if err := compileSvc.CompileStrategy(strategy.Path); err != nil {
//			return fmt.Errorf("failed to compile strategy: %w", err)
//		}
//	}
//
//	// Get the .so file path
//	strategyName := filepath.Base(strategy.Path)
//	soPath, err := filepath.Abs(filepath.Join(strategy.Path, strategyName+".so"))
//	if err != nil {
//		return fmt.Errorf("failed to get absolute path for strategy: %w", err)
//	}
//
//	// Final check if .so file exists (should exist after compilation)
//	if _, err := os.Stat(soPath); os.IsNotExist(err) {
//		return fmt.Errorf("strategy plugin not found after compilation: %s", soPath)
//	}
//
//	// Path to kronos-live binary
//	// TODO: Make this configurable or discover it automatically
//	kronosLivePath := "/Users/williamr/Documents/holdex/repos/live-trading/bin/kronos-live"
//
//	// Check if kronos-live exists
//	if _, err := os.Stat(kronosLivePath); os.IsNotExist(err) {
//		return fmt.Errorf("kronos-live binary not found: %s", kronosLivePath)
//	}
//
//	// Build command arguments
//	args := []string{"run", "--exchange", exchange.Name, "--strategy", soPath}
//
//	// Add exchange-specific flags
//	switch exchange.Name {
//	case "paradex":
//		if accountAddr, ok := exchange.Credentials["account_address"]; ok && accountAddr != "" {
//			args = append(args, "--paradex-account-address", accountAddr)
//		} else {
//			return fmt.Errorf("paradex account address is required")
//		}
//
//		if ethKey, ok := exchange.Credentials["eth_private_key"]; ok && ethKey != "" {
//			args = append(args, "--paradex-eth-private-key", ethKey)
//		} else {
//			return fmt.Errorf("paradex eth private key is required")
//		}
//
//		if l2Key, ok := exchange.Credentials["l2_private_key"]; ok && l2Key != "" {
//			args = append(args, "--paradex-l2-private-key", l2Key)
//		}
//
//		if exchange.Network != "" {
//			args = append(args, "--paradex-network", exchange.Network)
//		}
//
//	case "bybit", "binance", "kraken":
//		if apiKey, ok := exchange.Credentials["api_key"]; ok && apiKey != "" {
//			args = append(args, "--api-key", apiKey)
//		} else {
//			return fmt.Errorf("%s api key is required", exchange.Name)
//		}
//
//		if apiSecret, ok := exchange.Credentials["api_secret"]; ok && apiSecret != "" {
//			args = append(args, "--api-secret", apiSecret)
//		} else {
//			return fmt.Errorf("%s api secret is required", exchange.Name)
//		}
//
//	default:
//		return fmt.Errorf("unsupported exchange: %s", exchange.Name)
//	}
//
//	// Create and execute the command
//	cmd := exec.Command(kronosLivePath, args...)
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//	cmd.Stdin = os.Stdin
//
//	// Run the command (this will block until the user stops it with Ctrl+C)
//	return cmd.Run()
//}

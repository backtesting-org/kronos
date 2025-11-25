package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CompileService handles compilation of strategies into .so plugins
type CompileService struct{}

func NewCompileService() *CompileService {
	return &CompileService{}
}

// CompileStrategy compiles a strategy's .go file into a .so plugin if needed
func (s *CompileService) CompileStrategy(strategyPath string) error {
	strategyName := filepath.Base(strategyPath)
	strategyGoPath := filepath.Join(strategyPath, "strategy.go")
	soPath := filepath.Join(strategyPath, strategyName+".so")

	// Check if strategy.go exists
	if _, err := os.Stat(strategyGoPath); os.IsNotExist(err) {
		return fmt.Errorf("strategy.go not found in %s", strategyPath)
	}

	// Check if .so exists and is up-to-date
	goInfo, err := os.Stat(strategyGoPath)
	if err != nil {
		return err
	}

	soInfo, err := os.Stat(soPath)
	if err == nil && soInfo.ModTime().After(goInfo.ModTime()) {
		// .so exists and is newer than .go - no need to rebuild
		return nil
	}

	// Need to compile
	fmt.Printf("🔨 Compiling %s strategy...\n", strategyName)

	// First, run go mod tidy to ensure all dependencies are downloaded
	fmt.Printf("  📦 Downloading dependencies...\n")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = strategyPath
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("failed to download dependencies for %s: %w", strategyName, err)
	}

	// Now compile the plugin
	fmt.Printf("  🔧 Building plugin...\n")
	// Use relative paths since we're setting cmd.Dir to strategyPath
	outputFileName := strategyName + ".so"
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFileName, "strategy.go")
	cmd.Dir = strategyPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile %s: %w", strategyName, err)
	}

	fmt.Printf("✅ Compiled %s.so successfully\n\n", strategyName)
	return nil
}

// PreCompileStrategies scans and compiles all strategies in the strategies directory
func (s *CompileService) PreCompileStrategies(strategiesDir string) {
	// Check if strategies directory exists
	entries, err := os.ReadDir(strategiesDir)
	if err != nil {
		return // No strategies directory, skip
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		strategyPath := filepath.Join(strategiesDir, entry.Name())
		configPath := filepath.Join(strategyPath, "config.yml")

		// Only compile if config.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Try to compile (errors are printed but not fatal)
		s.CompileStrategy(strategyPath)
	}
}

// IsCompiled checks if a strategy has a compiled .so file
func (s *CompileService) IsCompiled(strategyPath string) bool {
	strategyName := filepath.Base(strategyPath)
	soPath := filepath.Join(strategyPath, strategyName+".so")
	_, err := os.Stat(soPath)
	return err == nil
}

// NeedsRecompile checks if a strategy needs to be recompiled
func (s *CompileService) NeedsRecompile(strategyPath string) bool {
	strategyName := filepath.Base(strategyPath)
	strategyGoPath := filepath.Join(strategyPath, "strategy.go")
	soPath := filepath.Join(strategyPath, strategyName+".so")

	goInfo, err := os.Stat(strategyGoPath)
	if err != nil {
		return true // Can't stat .go file, assume needs recompile
	}

	soInfo, err := os.Stat(soPath)
	if err != nil {
		return true // .so doesn't exist, needs compile
	}

	// Check if .go is newer than .so
	return goInfo.ModTime().After(soInfo.ModTime())
}

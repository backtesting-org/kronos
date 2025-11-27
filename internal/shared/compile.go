package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type CompileService interface {
	CompileStrategy(strategyPath string) error
	PreCompileStrategies(strategiesDir string) map[string]error
	IsCompiled(strategyPath string) bool
	NeedsRecompile(strategyPath string) bool
}

// compileService handles compilation of strategies into .so plugins
type compileService struct{}

func NewCompileService() CompileService {
	return &compileService{}
}

// CompileStrategy compiles a strategy's .go file into a .so plugin if needed
func (s *compileService) CompileStrategy(strategyPath string) error {
	strategyName := filepath.Base(strategyPath)
	strategyGoPath := filepath.Join(strategyPath, "strategy.go")
	soPath := filepath.Join(strategyPath, strategyName+".so")

	// Check if strategy.go exists
	if _, err := os.Stat(strategyGoPath); os.IsNotExist(err) {
		return fmt.Errorf("strategy.go not found")
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
	tidyOutput, err := tidyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to download dependencies: %s", string(tidyOutput))
	}

	// Now compile the plugin
	fmt.Printf("  🔧 Building plugin...\n")
	// Use relative paths since we're setting cmd.Dir to strategyPath
	outputFileName := strategyName + ".so"
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFileName, "strategy.go")
	cmd.Dir = strategyPath

	// Capture both stdout and stderr to show detailed error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return the compilation error with full output
		return fmt.Errorf("compilation failed: %s", string(output))
	}

	fmt.Printf("✅ Compiled %s.so successfully\n\n", strategyName)
	return nil
}

// PreCompileStrategies scans and compiles all strategies in the strategies directory
// Returns a map of strategy names to compilation errors (if any)
func (s *compileService) PreCompileStrategies(strategiesDir string) map[string]error {
	errors := make(map[string]error)

	// Check if strategies directory exists
	entries, err := os.ReadDir(strategiesDir)
	if err != nil {
		return errors // No strategies directory, return empty map
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		strategyName := entry.Name()
		strategyPath := filepath.Join(strategiesDir, strategyName)
		configPath := filepath.Join(strategyPath, "config.yml")

		// Only compile if config.yml exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue
		}

		// Try to compile and capture any errors
		if err := s.CompileStrategy(strategyPath); err != nil {
			errors[strategyName] = err
		}
	}

	return errors
}

// IsCompiled checks if a strategy has a compiled .so file
func (s *compileService) IsCompiled(strategyPath string) bool {
	strategyName := filepath.Base(strategyPath)
	soPath := filepath.Join(strategyPath, strategyName+".so")
	_, err := os.Stat(soPath)
	return err == nil
}

// NeedsRecompile checks if a strategy needs to be recompiled
func (s *compileService) NeedsRecompile(strategyPath string) bool {
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

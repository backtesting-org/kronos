package scaffold

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/afero"
)

type Scaffolder struct {
	fs afero.Fs
}

func NewScaffolder() *Scaffolder {
	return &Scaffolder{
		fs: afero.NewOsFs(),
	}
}

type ProjectData struct {
	ProjectName     string
	ModulePath      string
	StrategyPackage string
}

func (s *Scaffolder) CreateProject(name string) error {
	return s.CreateProjectWithStrategy(name, "mean_reversion")
}

func (s *Scaffolder) CreateProjectWithStrategy(name, strategyExample string) error {
	green := color.New(color.FgGreen, color.Bold)
	fmt.Printf("🚀 Creating Kronos project: %s\n\n", green.Sprint(name))

	// Check if exists
	if exists, _ := afero.DirExists(s.fs, name); exists {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	data := ProjectData{
		ProjectName:     name,
		ModulePath:      "github.com/your-username/" + name,
		StrategyPackage: strategyExample,
	}

	// Generate files (git clone will create the directory)
	if err := s.generateFiles(name, strategyExample, data); err != nil {
		return err
	}

	s.printSuccess(name)
	return nil
}

func (s *Scaffolder) generateFiles(name, strategyExample string, data ProjectData) error {
	// Git clone with sparse checkout directly to project directory
	fmt.Printf("  📦 Downloading %s example from GitHub...\n", strategyExample)

	cmd := exec.Command("git", "clone", "--depth", "1", "--filter=blob:none", "--sparse",
		"https://github.com/backtesting-org/kronos-sdk.git", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone SDK: %w", err)
	}

	// Set sparse checkout to get ONLY the selected example
	examplePath := fmt.Sprintf("examples/%s", strategyExample)
	cmd = exec.Command("git", "-C", name, "sparse-checkout", "set", examplePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", strategyExample, err)
	}

	// Move everything from examples/{strategy} to root
	strategyDir := filepath.Join(name, "examples", strategyExample)
	files, err := os.ReadDir(strategyDir)
	if err != nil {
		return fmt.Errorf("failed to read %s directory: %w", strategyExample, err)
	}

	for _, file := range files {
		srcPath := filepath.Join(strategyDir, file.Name())
		dstPath := filepath.Join(name, file.Name())

		if err := os.Rename(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", file.Name(), err)
		}
		fmt.Printf("  📝 %s\n", file.Name())
	}

	// Remove examples directory
	if err := os.RemoveAll(filepath.Join(name, "examples")); err != nil {
		return fmt.Errorf("failed to remove examples directory: %w", err)
	}

	// Remove .git directory
	if err := os.RemoveAll(filepath.Join(name, ".git")); err != nil {
		return fmt.Errorf("failed to remove .git directory: %w", err)
	}

	// Generate configuration files
	if err := s.generateConfigFiles(name, strategyExample); err != nil {
		return fmt.Errorf("failed to generate config files: %w", err)
	}

	return nil
}

func (s *Scaffolder) generateConfigFiles(name, strategyExample string) error {
	// Note: config.yml comes from the SDK example and contains only metadata
	// We do NOT generate it here - it's downloaded with the strategy

	// Generate exchanges.yml with assets configuration
	exchangesYAML := `# Global Exchange Configuration
# Configure which exchanges and assets to trade

exchanges:
  - name: binance
    enabled: true
    credentials:
      api_key: ""
      api_secret: ""
    assets:
      - BTC/USDT
      - ETH/USDT

  - name: bybit
    enabled: true
    credentials:
      api_key: ""
      api_secret: ""
    assets:
      - BTC/USDT

  - name: paradex
    enabled: false
    credentials:
      account_address: ""
      eth_private_key: ""
    assets:
      - BTC/USD
`

	exchangesPath := filepath.Join(name, "exchanges.yml")
	if err := os.WriteFile(exchangesPath, []byte(exchangesYAML), 0644); err != nil {
		return fmt.Errorf("failed to write exchanges.yml: %w", err)
	}
	fmt.Printf("  📝 exchanges.yml\n")

	// Generate .gitignore if it doesn't exist
	gitignorePath := filepath.Join(name, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := `# Credentials
exchanges.yml

# Build artifacts
*.so
bin/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
`
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to write .gitignore: %w", err)
		}
		fmt.Printf("  📝 .gitignore\n")
	}

	return nil
}

func (s *Scaffolder) printSuccess(name string) {
	blue := color.New(color.FgBlue)

	fmt.Printf("\n✅ Project created!\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  %s\n", blue.Sprint("cd "+name))
	fmt.Printf("  %s\n", blue.Sprint("go mod tidy"))
	fmt.Printf("  %s\n", blue.Sprint("go run strategy.go"))
	fmt.Printf("\n")
	fmt.Printf("📝 Important:\n")
	fmt.Printf("  • exchanges.yml - Add your API credentials\n")
}

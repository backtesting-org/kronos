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
	green := color.New(color.FgGreen, color.Bold)
	fmt.Printf("🚀 Creating Kronos project: %s\n\n", green.Sprint(name))

	// Check if exists
	if exists, _ := afero.DirExists(s.fs, name); exists {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	data := ProjectData{
		ProjectName:     name,
		ModulePath:      "github.com/your-username/" + name,
		StrategyPackage: "example",
	}

	// Generate files (git clone will create the directory)
	if err := s.generateFiles(name, data); err != nil {
		return err
	}

	s.printSuccess(name)
	return nil
}

func (s *Scaffolder) generateFiles(name string, data ProjectData) error {
	// Git clone with sparse checkout directly to project directory
	fmt.Println("  📦 Downloading cash_carry example from GitHub...")

	cmd := exec.Command("git", "clone", "--depth", "1", "--filter=blob:none", "--sparse",
		"https://github.com/backtesting-org/kronos-sdk.git", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone SDK: %w", err)
	}

	// Set sparse checkout to get ONLY cash_carry
	cmd = exec.Command("git", "-C", name, "sparse-checkout", "set", "examples/cash_carry")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout cash_carry: %w", err)
	}

	// Move everything from examples/cash_carry to root
	cashCarryDir := filepath.Join(name, "examples", "cash_carry")
	files, err := os.ReadDir(cashCarryDir)
	if err != nil {
		return fmt.Errorf("failed to read cash_carry directory: %w", err)
	}

	for _, file := range files {
		srcPath := filepath.Join(cashCarryDir, file.Name())
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

	return nil
}

func (s *Scaffolder) printSuccess(name string) {
	blue := color.New(color.FgBlue)

	fmt.Printf("\n✅ Project created!\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  %s\n", blue.Sprint("cd "+name))
	fmt.Printf("  %s\n", blue.Sprint("go mod tidy"))
	fmt.Printf("  %s\n", blue.Sprint("go run strategy.go"))
}

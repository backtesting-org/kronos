package scaffold

import (
	"embed"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/fatih/color"
	"github.com/spf13/afero"
)

//go:embed templates/*
var templateFS embed.FS

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

	// Create structure
	if err := s.createDirs(name); err != nil {
		return err
	}

	// Generate files
	if err := s.generateFiles(name, data); err != nil {
		return err
	}

	s.printSuccess(name)
	return nil
}

func (s *Scaffolder) createDirs(name string) error {
	dirs := []string{
		name,
		filepath.Join(name, "strategies", "example"),
	}

	for _, dir := range dirs {
		if err := s.fs.MkdirAll(dir, 0755); err != nil {
			return err
		}
		fmt.Printf("  📁 %s\n", dir)
	}
	return nil
}

func (s *Scaffolder) generateFiles(name string, data ProjectData) error {
	files := map[string]string{
		"kronos.yml.tmpl":  "kronos.yml",
		"strategy.go.tmpl": "strategies/example/strategy.go",
		"go.mod.tmpl":      "go.mod",
		"gitignore.tmpl":   ".gitignore",
		"README.md.tmpl":   "README.md",
	}

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

	for src, dst := range files {
		dstPath := filepath.Join(name, dst)
		if err := s.renderTemplate(tmpl, src, dstPath, data); err != nil {
			return err
		}
		fmt.Printf("  📝 %s\n", dstPath)
	}

	return nil
}

func (s *Scaffolder) renderTemplate(tmpl *template.Template, name, path string, data ProjectData) error {
	file, err := s.fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.ExecuteTemplate(file, name, data)
}

func (s *Scaffolder) printSuccess(name string) {
	blue := color.New(color.FgBlue)

	fmt.Printf("\n✅ Project created!\n\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  %s\n", blue.Sprint("cd "+name))
	fmt.Printf("  %s\n", blue.Sprint("go mod tidy"))
	fmt.Printf("  %s\n", blue.Sprint("# Edit strategies/example/strategy.go"))
}

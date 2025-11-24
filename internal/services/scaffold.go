package services

import (
	"github.com/backtesting-org/kronos-cli/internal/scaffold"
)

// ScaffoldService handles project creation and scaffolding
type ScaffoldService struct {
	scaffolder *scaffold.Scaffolder
}

func NewScaffoldService() *ScaffoldService {
	return &ScaffoldService{
		scaffolder: scaffold.NewScaffolder(),
	}
}

func (s *ScaffoldService) CreateProject(name string) error {
	return s.scaffolder.CreateProject(name)
}

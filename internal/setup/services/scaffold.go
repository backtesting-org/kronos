package services

import (
	"github.com/backtesting-org/kronos-cli/internal/setup/scaffold"
	"github.com/backtesting-org/kronos-cli/internal/setup/types"
)

// scaffoldService handles project creation and scaffolding
type scaffoldService struct {
	scaffolder scaffold.Scaffolder
}

func NewScaffoldService(scaffolder scaffold.Scaffolder) types.ScaffoldService {
	return &scaffoldService{
		scaffolder: scaffolder,
	}
}

func (s *scaffoldService) CreateProject(name string) error {
	return s.scaffolder.CreateProject(name)
}

func (s *scaffoldService) CreateProjectWithStrategy(name, strategyExample string) error {
	return s.scaffolder.CreateProjectWithStrategy(name, strategyExample)
}

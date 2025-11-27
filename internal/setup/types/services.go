package types

type ScaffoldService interface {
	CreateProject(name string) error
	CreateProjectWithStrategy(name, strategyExample string) error
}

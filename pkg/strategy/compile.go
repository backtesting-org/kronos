package strategy

type CompileService interface {
	CompileStrategy(strategyPath string) error
	PreCompileStrategies(strategiesDir string) map[string]error
	IsCompiled(strategyPath string) bool
	NeedsRecompile(strategyPath string) bool
}

package interactive

// Screen represents which screen we're on
type Screen int

const (
	ScreenSelection Screen = iota
	ScreenExchangeSelection
	ScreenCredentials
	ScreenConfirmation
	ScreenDeploying
	ScreenSuccess
	ScreenEmptyState
)

const (
	// visibleStrategies is the maximum number of strategies shown at once
	visibleStrategies = 3
)

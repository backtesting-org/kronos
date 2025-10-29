package scaffold

type StrategyTemplateData struct {
	ProjectName     string
	StrategyName    string
	StrategyPackage string
	StrategyType    string
	Description     string
	RiskLevel       string
	Parameters      []ParameterData
}

type ParameterData struct {
	Name         string
	Type         string
	DefaultValue string
	Comment      string
}

var StrategyTemplates = map[string]StrategyTemplateData{
	"moving-average": {
		StrategyName:    "MovingAverage",
		StrategyPackage: "moving_average",
		StrategyType:    "MovingAverage",
		Description:     "Moving average crossover strategy",
		RiskLevel:       "RiskLevelMedium",
		Parameters: []ParameterData{
			{
				Name:         "shortPeriod",
				Type:         "int",
				DefaultValue: "10",
				Comment:      "Short MA period",
			},
			{
				Name:         "longPeriod",
				Type:         "int",
				DefaultValue: "30",
				Comment:      "Long MA period",
			},
		},
	},
	"cash-carry": {
		StrategyName:    "CashCarry",
		StrategyPackage: "cash_carry",
		StrategyType:    "CashCarry",
		Description:     "Cash carry arbitrage strategy",
		RiskLevel:       "RiskLevelLow",
		Parameters: []ParameterData{
			{
				Name:         "minFundingRate",
				Type:         "decimal.Decimal",
				DefaultValue: "decimal.RequireFromString(\"0.001\")",
				Comment:      "Minimum funding rate threshold",
			},
		},
	},
}

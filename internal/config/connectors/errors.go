package connectors

import (
	"fmt"

	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	localConnector "github.com/backtesting-org/live-trading/pkg/connector"
)

// ValidationError provides detailed information about what went wrong during connector validation
type ValidationError struct {
	Exchange         string            // The exchange name
	Missing          []string          // Missing required fields
	InvalidFields    map[string]string // Field name -> reason why invalid
	ExchangeNotFound bool              // Is the exchange not in the registry?
	NotEnabled       bool              // Is the exchange disabled in config?
	MappingError     string            // Error during config mapping
	SDKValidationErr string            // Error from SDK's validation
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.ExchangeNotFound {
		return fmt.Sprintf("exchange '%s' is not available in the SDK registry", ve.Exchange)
	}

	if ve.NotEnabled {
		return fmt.Sprintf("exchange '%s' is configured but not enabled in exchanges.yml", ve.Exchange)
	}

	if len(ve.Missing) > 0 {
		return fmt.Sprintf("exchange '%s' is missing required credentials: %v", ve.Exchange, ve.Missing)
	}

	if len(ve.InvalidFields) > 0 {
		details := ""
		for field, reason := range ve.InvalidFields {
			details += fmt.Sprintf("\n  • %s: %s", field, reason)
		}
		return fmt.Sprintf("exchange '%s' has invalid fields:%s", ve.Exchange, details)
	}

	if ve.MappingError != "" {
		return fmt.Sprintf("exchange '%s' config mapping failed: %s", ve.Exchange, ve.MappingError)
	}

	if ve.SDKValidationErr != "" {
		return fmt.Sprintf("exchange '%s' configuration is invalid: %s", ve.Exchange, ve.SDKValidationErr)
	}

	return fmt.Sprintf("exchange '%s' validation failed for unknown reason", ve.Exchange)
}

// IsExchangeNotFound returns true if the exchange isn't in the registry
func (ve *ValidationError) IsExchangeNotFound() bool {
	return ve.ExchangeNotFound
}

// IsNotEnabled returns true if the exchange is not enabled
func (ve *ValidationError) IsNotEnabled() bool {
	return ve.NotEnabled
}

// IsMissingCredentials returns true if required credentials are missing
func (ve *ValidationError) IsMissingCredentials() bool {
	return len(ve.Missing) > 0
}

// IsInvalidConfig returns true if the config is invalid
func (ve *ValidationError) IsInvalidConfig() bool {
	return len(ve.InvalidFields) > 0 || ve.SDKValidationErr != ""
}

// StrategyValidationError contains validation results for all exchanges in a strategy
type StrategyValidationError struct {
	Strategy            string                                           // Strategy name
	ExchangeNames       []string                                         // Requested exchanges
	SuccessfulExchanges map[connector.ExchangeName]localConnector.Config // Exchanges that loaded successfully
	ValidationErrors    map[string]*ValidationError                      // Exchange name -> validation error
}

// Error implements the error interface
func (sve *StrategyValidationError) Error() string {
	msg := fmt.Sprintf("strategy validation failed for exchanges: %v\n\n", sve.ExchangeNames)

	if len(sve.ValidationErrors) == 0 {
		return msg + "all exchanges validated successfully"
	}

	msg += fmt.Sprintf("Failed exchanges: %d/%d\n\n", len(sve.ValidationErrors), len(sve.ExchangeNames))

	for exName, valErr := range sve.ValidationErrors {
		msg += fmt.Sprintf("❌ %s:\n", exName)
		if valErr.ExchangeNotFound {
			msg += "   Not found in SDK registry\n"
		} else if valErr.NotEnabled {
			msg += "   Not enabled in exchanges.yml\n"
		} else if len(valErr.Missing) > 0 {
			msg += fmt.Sprintf("   Missing credentials: %v\n", valErr.Missing)
		} else if len(valErr.InvalidFields) > 0 {
			for field, reason := range valErr.InvalidFields {
				msg += fmt.Sprintf("   Invalid %s: %s\n", field, reason)
			}
		} else if valErr.MappingError != "" {
			msg += fmt.Sprintf("   Mapping error: %s\n", valErr.MappingError)
		} else if valErr.SDKValidationErr != "" {
			msg += fmt.Sprintf("   SDK validation: %s\n", valErr.SDKValidationErr)
		}
	}

	return msg
}

// GetExchangeError returns the validation error for a specific exchange
func (sve *StrategyValidationError) GetExchangeError(exchangeName string) *ValidationError {
	return sve.ValidationErrors[exchangeName]
}

// HasExchangeError checks if a specific exchange has a validation error
func (sve *StrategyValidationError) HasExchangeError(exchangeName string) bool {
	_, exists := sve.ValidationErrors[exchangeName]
	return exists
}

// GetExchangesByProblem returns all exchanges that have a specific type of problem
func (sve *StrategyValidationError) GetExchangesByProblem(problemType string) []string {
	var result []string

	for exName, valErr := range sve.ValidationErrors {
		switch problemType {
		case "not_found":
			if valErr.ExchangeNotFound {
				result = append(result, exName)
			}
		case "not_enabled":
			if valErr.NotEnabled {
				result = append(result, exName)
			}
		case "missing_credentials":
			if len(valErr.Missing) > 0 {
				result = append(result, exName)
			}
		case "invalid_config":
			if len(valErr.InvalidFields) > 0 || valErr.SDKValidationErr != "" {
				result = append(result, exName)
			}
		case "mapping_error":
			if valErr.MappingError != "" {
				result = append(result, exName)
			}
		}
	}

	return result
}

// SuccessCount returns how many exchanges validated successfully
func (sve *StrategyValidationError) SuccessCount() int {
	return len(sve.SuccessfulExchanges)
}

// FailureCount returns how many exchanges failed validation
func (sve *StrategyValidationError) FailureCount() int {
	return len(sve.ValidationErrors)
}

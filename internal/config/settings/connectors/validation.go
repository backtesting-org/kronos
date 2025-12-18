package connectors

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/live-trading/pkg/connectors"
)

// ValidationService handles validation of settings and connectors
type ValidationService interface {
	ValidateConnectorName(name string) error
	ValidateUniqueNames(connectors []settings.Connector) error
	ValidateSettings(config *settings.Settings) error
	GetAvailableConnectors() []connector.ExchangeName
}

type validationService struct {
	// No dependencies - validation is self-contained
}

func NewValidationService() ValidationService {
	return &validationService{}
}

// GetAvailableConnectors returns list of connectors available in SDK
func (v *validationService) GetAvailableConnectors() []connector.ExchangeName {
	return connectors.ListAvailable()
}

// ValidateConnectorName checks if the connector name is available in the SDK
func (v *validationService) ValidateConnectorName(name string) error {
	available := v.GetAvailableConnectors()

	for _, exchangeName := range available {
		if string(exchangeName) == name {
			return nil
		}
	}

	return fmt.Errorf("connector '%s' is not available in the SDK", name)
}

// ValidateUniqueNames ensures no duplicate connector names exist
func (v *validationService) ValidateUniqueNames(connectorList []settings.Connector) error {
	seen := make(map[string]bool)

	for _, conn := range connectorList {
		if seen[conn.Name] {
			return fmt.Errorf("duplicate connector name: '%s'", conn.Name)
		}
		seen[conn.Name] = true
	}

	return nil
}

// ValidateSettings performs full validation on a Settings object
// Note: This only validates structure and availability, not credentials
// Credential validation happens when actually connecting to the exchange
func (v *validationService) ValidateSettings(config *settings.Settings) error {
	// Validate unique names
	if err := v.ValidateUniqueNames(config.Connectors); err != nil {
		return err
	}

	// Validate each connector
	for _, conn := range config.Connectors {
		// Validate name is available
		if err := v.ValidateConnectorName(conn.Name); err != nil {
			return err
		}

		// Basic structure validation
		if conn.Name == "" {
			return fmt.Errorf("connector name cannot be empty")
		}

		// Validate credentials are present (not empty)
		if conn.Enabled && len(conn.Credentials) == 0 {
			return fmt.Errorf("connector '%s' is enabled but has no credentials configured", conn.Name)
		}
	}

	return nil
}

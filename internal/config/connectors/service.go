package connectors

import (
	"encoding/json"
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/config/settings"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	localConnector "github.com/backtesting-org/live-trading/pkg/connector"
	"github.com/backtesting-org/live-trading/pkg/connectors"
)

type ConnectorService interface {
	FetchAvailableConnectors() []connector.ExchangeName
	GetMatchingConnectors() (map[connector.ExchangeName]settings.Connector, error)
	ValidateConnectorConfig(exchangeName connector.ExchangeName, userConnector settings.Connector) error
	MapToSDKConfig(userConnector settings.Connector) (localConnector.Config, error)
	GetConnectorConfigsForStrategy(exchangeNames []string) (map[connector.ExchangeName]localConnector.Config, error)
}

type connectorService struct {
	config settings.Configuration
}

func NewConnectorService(config settings.Configuration) ConnectorService {
	return &connectorService{
		config: config,
	}
}

func (c *connectorService) FetchAvailableConnectors() []connector.ExchangeName {
	return connectors.ListAvailable()
}

// GetMatchingConnectors returns user-configured connectors that are also available in the SDK
func (c *connectorService) GetMatchingConnectors() (map[connector.ExchangeName]settings.Connector, error) {
	// Get available connectors from SDK
	availableConnectors := c.FetchAvailableConnectors()

	// Create a lookup map for quick checking
	availableMap := make(map[string]bool)
	for _, exchangeName := range availableConnectors {
		availableMap[string(exchangeName)] = true
	}

	// Get user's configured connectors from the settings service
	userConnectors, err := c.config.GetConnectors()
	if err != nil {
		return nil, err
	}

	// Filter to only return matching connectors as a map
	matchingConnectors := make(map[connector.ExchangeName]settings.Connector)
	for _, conn := range userConnectors {
		if availableMap[conn.Name] {
			matchingConnectors[connector.ExchangeName(conn.Name)] = conn
		}
	}

	return matchingConnectors, nil
}

// ValidateConnectorConfig validates if a specific exchange has the right configuration loaded
func (c *connectorService) ValidateConnectorConfig(exchangeName connector.ExchangeName, userConnector settings.Connector) error {
	ve := &ValidationError{
		Exchange:      string(exchangeName),
		Missing:       []string{},
		InvalidFields: make(map[string]string),
	}

	// Check if the connector is available in SDK
	if !connectors.IsAvailable(exchangeName) {
		ve.ExchangeNotFound = true
		return ve
	}

	// Check if the connector is enabled
	if !userConnector.Enabled {
		ve.NotEnabled = true
		return ve
	}

	// Check if the user connector name matches the exchange name
	if userConnector.Name != string(exchangeName) {
		ve.InvalidFields["name"] = fmt.Sprintf("expected '%s', got '%s'", exchangeName, userConnector.Name)
		return ve
	}

	// Map user connector to SDK config
	sdkConfig, err := c.MapToSDKConfig(userConnector)
	if err != nil {
		ve.MappingError = err.Error()
		return ve
	}

	// Validate the SDK config using the SDK's own validation logic
	if err := sdkConfig.Validate(); err != nil {
		ve.SDKValidationErr = err.Error()
		return ve
	}

	return nil
}

// MapToSDKConfig maps a user connector configuration to the appropriate SDK config type
// This uses the SDK's config templates and generically maps the user's credentials
func (c *connectorService) MapToSDKConfig(userConnector settings.Connector) (localConnector.Config, error) {
	exchangeName := connector.ExchangeName(userConnector.Name)

	// Get the config type template for this exchange from the SDK
	configTemplate := connectors.GetConfigType(exchangeName)
	if configTemplate == nil {
		return nil, fmt.Errorf("no config template found for exchange '%s'", exchangeName)
	}

	// Create a map to hold all the user's configuration data
	configData := make(map[string]interface{})

	// Copy credentials
	for key, value := range userConnector.Credentials {
		configData[key] = value
	}

	// Add network-related fields if present
	if userConnector.Network != "" {
		configData["network"] = userConnector.Network
		configData["use_testnet"] = userConnector.Network == "testnet"
	}

	// Marshal to JSON and unmarshal into the SDK config type
	// This lets the SDK's config struct handle the mapping and field names
	jsonData, err := json.Marshal(configData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config data: %w", err)
	}

	// Create a new instance of the config type
	// We need to get a pointer to a new instance, not use the template directly
	sdkConfig := connectors.GetConfigType(exchangeName)
	if err := json.Unmarshal(jsonData, &sdkConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into SDK config: %w", err)
	}

	return sdkConfig, nil
}

// GetConnectorConfigsForStrategy returns validated and mapped SDK configs for the given exchange names
// Returns a StrategyValidationError if there are problems so callers can inspect specific issues
func (c *connectorService) GetConnectorConfigsForStrategy(exchangeNames []string) (map[connector.ExchangeName]localConnector.Config, error) {
	// Get all matching connectors (available in SDK AND configured by user)
	allConnectors, err := c.GetMatchingConnectors()
	if err != nil {
		return nil, fmt.Errorf("failed to get connectors: %w", err)
	}

	// Track detailed problems for each exchange
	validationResults := make(map[string]*ValidationError)

	// Filter to only the exchanges this strategy needs and map to SDK configs
	connectorConfigs := make(map[connector.ExchangeName]localConnector.Config)

	for _, stratExchangeName := range exchangeNames {
		exchangeName := connector.ExchangeName(stratExchangeName)

		// Check if this exchange exists in SDK
		if !connectors.IsAvailable(exchangeName) {
			ve := &ValidationError{
				Exchange:         stratExchangeName,
				ExchangeNotFound: true,
			}
			validationResults[stratExchangeName] = ve
			continue
		}

		// Check if this exchange is configured by user
		userConn, exists := allConnectors[exchangeName]
		if !exists {
			ve := &ValidationError{
				Exchange:         stratExchangeName,
				ExchangeNotFound: true,
			}
			validationResults[stratExchangeName] = ve
			continue
		}

		// Check if enabled
		if !userConn.Enabled {
			ve := &ValidationError{
				Exchange:   stratExchangeName,
				NotEnabled: true,
			}
			validationResults[stratExchangeName] = ve
			continue
		}

		// Validate and map to SDK config
		if err := c.ValidateConnectorConfig(exchangeName, userConn); err != nil {
			// Store the validation error details
			if valErr, ok := err.(*ValidationError); ok {
				validationResults[stratExchangeName] = valErr
			} else {
				validationResults[stratExchangeName] = &ValidationError{
					Exchange:         stratExchangeName,
					SDKValidationErr: err.Error(),
				}
			}
			continue
		}

		sdkConfig, err := c.MapToSDKConfig(userConn)
		if err != nil {
			validationResults[stratExchangeName] = &ValidationError{
				Exchange:     stratExchangeName,
				MappingError: err.Error(),
			}
			continue
		}

		connectorConfigs[exchangeName] = sdkConfig
	}

	// If we have any problems, return a detailed error
	if len(validationResults) > 0 {
		return nil, &StrategyValidationError{
			Strategy:            "",
			ExchangeNames:       exchangeNames,
			SuccessfulExchanges: connectorConfigs,
			ValidationErrors:    validationResults,
		}
	}

	return connectorConfigs, nil
}

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
	// Check if the connector is available
	if !connectors.IsAvailable(exchangeName) {
		return fmt.Errorf("connector '%s' is not available", exchangeName)
	}

	// Check if the user connector name matches the exchange name
	if userConnector.Name != string(exchangeName) {
		return fmt.Errorf("connector name mismatch: expected '%s', got '%s'", exchangeName, userConnector.Name)
	}

	// Map user connector to SDK config
	sdkConfig, err := c.MapToSDKConfig(userConnector)
	if err != nil {
		return fmt.Errorf("failed to map connector config: %w", err)
	}

	// Validate the SDK config using the SDK's own validation logic
	if err := sdkConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration for '%s': %w", exchangeName, err)
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
// This encapsulates all the logic of matching, filtering, validating, and mapping connectors
func (c *connectorService) GetConnectorConfigsForStrategy(exchangeNames []string) (map[connector.ExchangeName]localConnector.Config, error) {
	// Get all matching connectors (available in SDK AND configured by user)
	allConnectors, err := c.GetMatchingConnectors()
	if err != nil {
		return nil, fmt.Errorf("failed to get connectors: %w", err)
	}

	// Filter to only the exchanges this strategy needs and map to SDK configs
	connectorConfigs := make(map[connector.ExchangeName]localConnector.Config)

	for _, stratExchangeName := range exchangeNames {
		exchangeName := connector.ExchangeName(stratExchangeName)

		// Check if this exchange is in our matching connectors and enabled
		userConn, exists := allConnectors[exchangeName]
		if !exists || !userConn.Enabled {
			continue
		}

		// Validate and map to SDK config
		if err := c.ValidateConnectorConfig(exchangeName, userConn); err != nil {
			return nil, fmt.Errorf("invalid connector config for %s: %w", stratExchangeName, err)
		}

		sdkConfig, err := c.MapToSDKConfig(userConn)
		if err != nil {
			return nil, fmt.Errorf("failed to map connector config for %s: %w", stratExchangeName, err)
		}

		connectorConfigs[exchangeName] = sdkConfig
	}

	if len(connectorConfigs) == 0 {
		return nil, fmt.Errorf("no enabled connectors found for exchanges: %v", exchangeNames)
	}

	return connectorConfigs, nil
}

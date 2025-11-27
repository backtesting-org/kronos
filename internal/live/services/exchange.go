package services

import (
	"fmt"

	"github.com/backtesting-org/kronos-cli/internal/live/types"
	"github.com/spf13/viper"
)

// configService handles loading and validating kronos.yml
type configService struct {
	configPath string
}

func NewConfigService(configPath string) types.ConfigService {
	return &configService{
		configPath: configPath,
	}
}

// LoadExchangeCredentials loads exchange configurations from kronos.yml
func (s *configService) LoadExchangeCredentials() (types.Connectors, error) {
	v := viper.New()
	v.SetConfigFile(s.configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return types.Connectors{}, fmt.Errorf("failed to read %s: %w", s.configPath, err)
	}

	// Get exchanges section
	exchangesRaw := v.Get("exchanges")
	if exchangesRaw == nil {
		return types.Connectors{}, nil
	}

	// Parse as array of exchange configs
	exchangesList, ok := exchangesRaw.([]interface{})
	if !ok {
		return types.Connectors{}, fmt.Errorf("invalid format for exchanges in %s", s.configPath)
	}

	var connectors types.Connectors
	for _, exchRaw := range exchangesList {
		exchMap, ok := exchRaw.(map[string]interface{})
		if !ok {
			continue
		}

		var exchangeConfig types.ExchangeConfig

		// Parse name
		if name, ok := exchMap["name"].(string); ok {
			exchangeConfig.Name = name
		}

		// Parse enabled
		if enabled, ok := exchMap["enabled"].(bool); ok {
			exchangeConfig.Enabled = enabled
		}

		// Parse credentials
		if credsRaw, ok := exchMap["credentials"].(map[string]interface{}); ok {
			exchangeConfig.Credentials = make(map[string]string)
			for key, val := range credsRaw {
				exchangeConfig.Credentials[key] = fmt.Sprint(val)
			}
		}

		// Parse assets
		if assetsRaw, ok := exchMap["assets"].([]interface{}); ok {
			for _, asset := range assetsRaw {
				if assetStr, ok := asset.(string); ok {
					exchangeConfig.Assets = append(exchangeConfig.Assets, assetStr)
				}
			}
		}

		connectors.Exchanges = append(connectors.Exchanges, exchangeConfig)
	}

	return connectors, nil
}

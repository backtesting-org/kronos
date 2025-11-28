package settings

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type settings struct {
	SettingsPath string
	settings     *Settings
}

func NewConfiguration() Configuration {
	return &settings{
		SettingsPath: "kronos.yml",
	}
}

// LoadSettings loads the Settings settings from the specified file path
func (c *settings) LoadSettings() (*Settings, error) {
	if c.settings != nil {
		return c.settings, nil
	}

	if !c.fileExists(c.SettingsPath) {
		return nil, fmt.Errorf("settings instance not found, please run 'settings init' to create one")
	}

	v := viper.New()
	v.SetConfigFile(c.SettingsPath)
	v.SetConfigType("yaml")

	// Enable environment variable substitution
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var settings Settings
	if err := v.Unmarshal(&settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Cache the config
	c.settings = &settings

	return c.settings, nil
}

// GetConnectors returns the cached exchange credentials from settings.yml
// If not loaded yet, it will load the settings config first
func (c *settings) GetConnectors() ([]Connector, error) {
	if c.settings != nil {
		return c.settings.Connectors, nil
	}

	// Load the full config which will also cache credentials
	if _, err := c.LoadSettings(); err != nil {
		return nil, err
	}

	return c.settings.Connectors, nil
}

// GetEnabledConnectors returns all enabled connectors
func (c *settings) GetEnabledConnectors() ([]Connector, error) {
	if c.settings == nil {
		// Load the full config which will also cache settings
		if _, err := c.LoadSettings(); err != nil {
			return nil, err
		}
	}

	enabled := make([]Connector, 0)
	for _, ex := range c.settings.Connectors {
		if ex.Enabled {
			enabled = append(enabled, ex)
		}
	}

	return enabled, nil
}

// FileExists checks if the config file exists
func (c *settings) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Locale             string `json:"locale,omitempty"`
	CurrencySymbol     string `json:"currency_symbol,omitempty"`
	DecimalSeparator   string `json:"decimal_separator,omitempty"`
	ThousandsSeparator string `json:"thousands_separator,omitempty"`
	Frequency          string `json:"frequency,omitempty"`
	Encrypted          bool   `json:"encrypted,omitempty"`
}

func ReadConfig(dbPath string) (Config, error) {
	configPath := filepath.Join(filepath.Dir(dbPath), "balde.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config not found")
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return config, nil
}

func WriteConfig(dbPath string, config Config) error {
	configPath := filepath.Join(filepath.Dir(dbPath), "balde.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

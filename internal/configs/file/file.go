package file

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func ConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".gophkeeper")

	info, err := os.Stat(configDir)
	if err == nil {
		if !info.IsDir() {
			if err := os.Remove(configDir); err != nil {
				return "", err
			}
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				return "", err
			}
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return "", err
		}
	} else {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig reads the config file and unmarshals it into a map.
// Returns an empty map if the config file does not exist.
func LoadConfig() (map[string]string, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]string), nil // no config file, return empty map
		}
		return nil, err
	}

	config := make(map[string]string)
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig marshals the config map and writes it to the config file.
func SaveConfig(config map[string]string) error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// GetConfigValue returns the value for a key from the config.
// Returns false if the key is not found.
func GetConfigValue(key string) (string, bool, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", false, err
	}
	val, ok := cfg[key]
	return val, ok, nil
}

// SetConfigValue sets a key to a value and saves the config.
func SetConfigValue(key, value string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg[key] = value
	return SaveConfig(cfg)
}

// UnsetConfigValue removes a key from the config and saves.
func UnsetConfigValue(key string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	delete(cfg, key)
	return SaveConfig(cfg)
}

// ListConfig returns a copy of all config key-value pairs.
func ListConfig() (map[string]string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	copyMap := make(map[string]string, len(cfg))
	for k, v := range cfg {
		copyMap[k] = v
	}
	return copyMap, nil
}

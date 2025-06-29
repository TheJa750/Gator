package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"dburl"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	path, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, configFileName), nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user

	err := write(*c)
	if err != nil {
		return err
	}

	return nil
}

func write(cfg Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func RestoreDefaultCfg() error {
	defaultCfg := Config{DbURL: "postgres://example"}

	err := write(defaultCfg)
	if err != nil {
		return err
	}

	return nil
}

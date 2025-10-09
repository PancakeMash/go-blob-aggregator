package config

import (
	"encoding/json"
	"os"
)

const configFilename = "/.gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	filePath, _ := getConfigFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func getConfigFilePath() (string, error) {
	Home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return Home + configFilename, nil
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const configFilePath = "./config.json"

type Config struct {
	Counter int `json:"counter"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		config.Counter = 0
		return config, nil
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

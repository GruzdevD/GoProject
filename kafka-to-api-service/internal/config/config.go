package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	KafkaBroker string `json:"kafka_broker"`
	APIEndpoint string `json:"api_endpoint"`
}

func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
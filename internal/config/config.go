package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
}

func LoadConfig() (*Config, error) {

	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil

}

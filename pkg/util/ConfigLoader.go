package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int    `yaml:"port"`
	Topic    string `yaml:"topic"`
	Server   string `yaml:"server"`
	Logging  bool   `yaml:"logging"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func (cfg *Config) PrintData() {
	// Print the config values for debugging purposes
	fmt.Println("CONFIG VALUES\n=======")
	fmt.Println("Port:", cfg.Port)
	fmt.Println("Topic:", cfg.Topic)
	fmt.Println("Server:", cfg.Server)
	fmt.Println("Logging:", cfg.Logging)
	fmt.Println("Name:", cfg.Name)
}

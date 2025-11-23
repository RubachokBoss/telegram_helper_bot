package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Postgress struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DB       string `yaml:"db"`
		SSLMode  string `yaml:"ssl_mode"`
	} `yaml:"postgres"`
	GRPC struct {
		Port string `yaml:"port"`
	} `yaml:"grpc"`
	Telegram struct {
		Token string `yaml:"token"`
	} `yaml:"telegram"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Postgres struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
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

	// Пытаемся загрузить из файла
	file, err := os.Open(configPath)
	if err == nil {
		defer file.Close()
		d := yaml.NewDecoder(file)
		if err := d.Decode(&config); err != nil {
			return nil, fmt.Errorf("cannot decode config: %v", err)
		}
	}

	// Переопределяем переменными окружения
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		config.Postgres.Host = host
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		config.Postgres.Port = port
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		config.Postgres.User = user
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		config.Postgres.Password = password
	}
	if dbname := os.Getenv("POSTGRES_DB"); dbname != "" {
		config.Postgres.DBName = dbname
	}

	// Для telegram-bot - полный адрес gRPC сервера
	if grpcHost := os.Getenv("GRPC_HOST"); grpcHost != "" {
		config.GRPC.Port = grpcHost
	}

	// Токен Telegram бота из переменных окружения
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		config.Telegram.Token = token
	}

	// Проверяем, что токен установлен
	if config.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	fmt.Printf("✅ Config loaded:\n")
	fmt.Printf("   DB Host: %s\n", config.Postgres.Host)
	fmt.Printf("   DB Port: %s\n", config.Postgres.Port)
	fmt.Printf("   DB Name: %s\n", config.Postgres.DBName)
	fmt.Printf("   DB User: %s\n", config.Postgres.User)
	fmt.Printf("   gRPC Port: %s\n", config.GRPC.Port)
	fmt.Printf("   Telegram Token: %s...\n", config.Telegram.Token[:10]) // Показываем только первые 10 символов

	return config, nil
}

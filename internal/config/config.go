package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string
		Port string
	}
	Logger struct {
		Level    string
		Encoding string
	}
}

func LoadConfig() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{
		Server: struct {
			Host string
			Port string
		}{
			Host: viper.GetString("SERVER_HOST"),
			Port: viper.GetString("SERVER_PORT"),
		},
		Logger: struct {
			Level    string
			Encoding string
		}{
			Level:    viper.GetString("LOG_LEVEL"),
			Encoding: viper.GetString("LOG_ENCODING"),
		},
	}

	return cfg, nil
}
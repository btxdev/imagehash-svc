package main

import (
	"log"

	"github.com/btxdev/imagehash-svc/internal/config"
	"github.com/btxdev/imagehash-svc/internal/server"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	srv := server.NewServer(logger)
	if err := srv.Start(cfg); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Logger.Encoding == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	err := config.Level.UnmarshalText([]byte(cfg.Logger.Level))
	if err != nil {
		return nil, err
	}

	return config.Build()
}
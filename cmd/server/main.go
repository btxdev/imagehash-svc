package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btxdev/imagehash-svc/internal/config"
	"github.com/btxdev/imagehash-svc/internal/imghash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"go.uber.org/zap"

	pb "github.com/btxdev/imagehash-svc/imagehash"
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

	hashHandler := imghash.NewImageHashHandler(logger)

	grpcServer := grpc.NewServer()
	pb.RegisterImagehashServiceServer(grpcServer, hashHandler)

	if cfg.Mode == "development" {
		reflection.Register(grpcServer)
		logger.Info("gRPC reflection enabled")
	}

	go func() {
		lis, err := net.Listen("tcp", net.JoinHostPort(cfg.GrpcServer.Host, cfg.GrpcServer.Port))
		if err != nil {
			logger.Fatal("Failed to listen", zap.Error(err))
		}

		logger.Info("Starting gRPC server", 
			zap.String("host", cfg.GrpcServer.Host),
			zap.String("port", cfg.GrpcServer.Port),
		)

		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	waitForShutdown(logger, grpcServer)
}

func waitForShutdown(logger *zap.Logger, grpcServer *grpc.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		logger.Warn("Server shutdown timed out, forcing exit")
		grpcServer.Stop()
	case <-stopped:
		logger.Info("Server stopped gracefully")
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
package server

import (
	"context"
	"net"

	"github.com/btxdev/imagehash-svc/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/btxdev/imagehash-svc/imagehash"
)

type Server struct {
	pb.ImagehashServiceServer
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Start(cfg *config.Config) error {
	lis, err := net.Listen("tcp", net.JoinHostPort(cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterImagehashServiceServer(grpcServer, s)

	s.logger.Info("Starting gRPC server",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port),
	)

	return grpcServer.Serve(lis)
}

func (s *Server) GetHash(ctx context.Context, req *pb.GetHashRequest) (*pb.GetHashResponse, error) {
	s.logger.Debug("GetHash RPC method called", zap.Any("request", req))
	a := req.GetA()
	b := req.GetB()
	c := a + b
	return &pb.GetHashResponse{
		C: c,
	}, nil
}
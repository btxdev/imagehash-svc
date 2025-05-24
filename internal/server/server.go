package server

import (
	"context"

	"go.uber.org/zap"

	pb "github.com/btxdev/imagehash-svc/imagehash"
)

type Server struct {
	pb.ImagehashServiceServer
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{logger: logger}
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
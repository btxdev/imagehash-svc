package server

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sync"

	_ "golang.org/x/image/webp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/btxdev/imagehash-svc/imagehash"
	"github.com/corona10/goimagehash"
)

type Server struct {
	pb.ImagehashServiceServer
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) GetHash(stream pb.ImagehashService_GetHashServer) error {
	var (
		meta      *pb.ImageMeta
		imageData bytes.Buffer
		mu        sync.Mutex
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		switch data := req.Data.(type) {
		case *pb.GetHashRequest_Meta:
			meta = data.Meta
			s.logger.Info("Receiving image",
				zap.String("filename", meta.Filename),
				zap.String("type", meta.MimeType),
				zap.Uint64("size", meta.FileSize),
			)

		case *pb.GetHashRequest_Chunk:
			mu.Lock()
			imageData.Write(data.Chunk.Content)
			mu.Unlock()

		default:
			return status.Error(codes.InvalidArgument, "unexpected message type")
		}
	}

	img, _, err := image.Decode(bytes.NewReader(imageData.Bytes()))
	if err != nil {
		s.logger.Error("Image decode failed", zap.Error(err))
		return status.Error(codes.InvalidArgument, "invalid image data")
	}

	hash, err := s.calculateHash(img)
	if err != nil {
		return status.Error(codes.Internal, "hash computation failed")
	}

	return stream.SendAndClose(&pb.GetHashResponse{
		Hash:           hash,
	})
}

func (s *Server) calculateHash(img image.Image) (string, error) {
	avgHash, err := goimagehash.AverageHash(img)
	if err != nil {
		return "", err
	}
	hash := avgHash.ToString()
	return hash, nil
}
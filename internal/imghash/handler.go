package imghash

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

type ImageHashHandler struct {
	pb.ImagehashServiceServer
	logger *zap.Logger
}

type imageHashResult struct {
	average string
	difference string
	perception string
}

func NewImageHashHandler(logger *zap.Logger) *ImageHashHandler {
	return &ImageHashHandler{logger: logger}
}

func (s *ImageHashHandler) GetHash(stream pb.ImagehashService_GetHashServer) error {
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

	hashes, err := s.calculateHashes(img, meta.GetHashConfig())
	if err != nil {
		return status.Error(codes.Internal, "hash computation failed")
	}

	return stream.SendAndClose(&pb.GetHashResponse{
		Average:           hashes.average,
		Difference:           hashes.difference,
		Perception:           hashes.perception,
	})
}

func (s *ImageHashHandler) calculateHashes(img image.Image, cfg *pb.HashConfig) (imageHashResult, error) {
	average := ""
	if cfg.Average {
		res, err := goimagehash.AverageHash(img)
		if err != nil {
			return imageHashResult{
				average: "",
				difference: "",
				perception: "",
			}, err
		}
		average = res.ToString()
	}

	difference := ""
	if cfg.Difference {
		res, err := goimagehash.DifferenceHash(img)
		if err != nil {
			return imageHashResult{
				average: average,
				difference: "",
				perception: "",
			}, err
		}
		difference = res.ToString()
	}

	perception := ""
	if cfg.Perception {
		res, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return imageHashResult{
				average: average,
				difference: difference,
				perception: "",
			}, err
		}
		perception = res.ToString()
	}

	return imageHashResult{
		average: average,
		difference: difference,
		perception: perception,
	}, nil
}
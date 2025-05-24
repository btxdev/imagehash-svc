package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	pb "github.com/btxdev/imagehash-svc/imagehash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewImagehashServiceClient(conn)

	filePath := "golang.png"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer file.Close()

	stat, _ := file.Stat()
	stream, err := client.GetHash(context.Background())
	if err != nil {
		log.Fatalf("GetHash stream error: %v", err)
	}

	err = stream.Send(&pb.GetHashRequest{
		Data: &pb.GetHashRequest_Meta{
			Meta: &pb.ImageMeta{
				Filename: filepath.Base(filePath),
				MimeType: "image/png",
				FileSize: uint64(stat.Size()),
			},
		},
	})
	if err != nil {
		log.Fatalf("meta send error: %v", err)
	}

	buf := make([]byte, 64*1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			break
		}

		err = stream.Send(&pb.GetHashRequest{
			Data: &pb.GetHashRequest_Chunk{
				Chunk: &pb.ImageChunk{
					Content: buf[:n],
				},
			},
		})
		if err != nil {
			log.Fatalf("chunk send error: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("CloseAndRecv error: %v", err)
	}

	log.Printf("Hash: %s", resp.Hash)
}
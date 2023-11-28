package main

import (
	"context"
	"crypto/rand"
	"flag"
	"log"
	"net"
	"time"

	pb "mocks3/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

var (
	port = flag.String("port", ":50051", "the port to listen on")
)

func (s *server) GetTimeToSleep(fileSize int64) time.Duration {
	// Wait based on file size
	return time.Duration(fileSize) * time.Millisecond
}

func (s *server) GetFile(ctx context.Context, in *pb.FileSize) (*pb.FileBlob, error) {
	// Wait based on file size
	log.Printf("Received a file size: %d", in.GetSize())
	time.Sleep(s.GetTimeToSleep(in.GetSize()))
	// Generate random blob
	blob := make([]byte, in.GetSize())
	rand.Read(blob)
	return &pb.FileBlob{Blob: blob}, nil
}

func (s *server) PutFile(ctx context.Context, in *pb.FileBlob) (*pb.FileSize, error) {
	// Wait
	log.Printf("Received a file blob of size: %d", len(in.GetBlob()))
	size := int64(len(in.GetBlob()))
	time.Sleep(s.GetTimeToSleep(size)) // predefined wait time
	return &pb.FileSize{Size: size}, nil
}

func main() {
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

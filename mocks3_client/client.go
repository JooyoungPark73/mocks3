package main

import (
	"context"
	"crypto/rand"
	"flag"
	"log"
	"time"

	pb "mocks3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Example 1: Send a file size and receive a random file
	size := int64(1024) // 1KB size
	r1, err := c.GetFile(ctx, &pb.FileSize{Size: size})
	if err != nil {
		if err == context.DeadlineExceeded {
			// handle the timeout
			log.Fatalf("request timed out: %v", err)
		} else {
			log.Fatalf("could not get file: %v", err)
		}
	}
	log.Printf("Received a file blob of size: %d", len(r1.GetBlob()))

	// Example 2: Send a random blob and get its size back
	// Create a random blob of data
	blob := make([]byte, size)
	rand.Read(blob)
	r2, err := c.PutFile(ctx, &pb.FileBlob{Blob: blob})
	if err != nil {
		log.Fatalf("could not get file size: %v", err)
	}
	log.Printf("Sent a blob, received size: %d", r2.GetSize())
}

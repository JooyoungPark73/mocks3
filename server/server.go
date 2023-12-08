package main

import (
	"context"
	"crypto/rand"
	"flag"
	"math"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	pb "github.com/JooyoungPark73/mocks3/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

var (
	port      = flag.String("port", "30000", "the port to listen on")
	verbosity = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
	buffer    = make([]byte, 512*1024*1024) // 512MB
)

func (s *server) GetTimeToSleep(commType string, fileSize int64) time.Duration {
	// Sourced from: https://github.com/vhive-serverless/MockS3/blob/main/mocks3/mock_io_functions.py
	// [  0.12018868   1.11999534 111.24820149]
	// a * np.exp(b * np.log10(x_point)) + c

	latencyPower := 0.12018868*math.Exp(1.11999534*math.Log10(float64(fileSize))) + 111.24820149

	if commType == "GET" {
		latencyPower = latencyPower * 0.67
	} else if commType == "PUT" {
		latencyPower = latencyPower * 1.33
	} else {
		log.Panic("Invalid communication type")
	}
	sleepTime := time.Duration(latencyPower) * time.Millisecond
	return sleepTime
}

func (s *server) GetFile(ctx context.Context, in *pb.FileSize) (*pb.FileBlob, error) {
	// Generate random blob
	size := in.GetSize()
	log.Debugf("Received a request for blob of size: %f KB", float64(size)/1024)
	blob := buffer[:size]

	return &pb.FileBlob{Blob: blob}, nil
}

func (s *server) PutFile(ctx context.Context, in *pb.FileBlob) (*pb.FileSize, error) {
	// Wait
	size := int64(len(in.GetBlob()))
	log.Debugf("Received a file blob of size: %f KB", float64(size)/1024)

	return &pb.FileSize{Size: size}, nil
}

func init() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)

	switch *verbosity {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	rand.Read(buffer)
}

func main() {
	maxMsgSize := int(math.Pow(2, 29)) // 512MB
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)
	pb.RegisterFileServiceServer(s, &server{})
	log.Infof("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

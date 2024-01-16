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

	latency := 120.18868*math.Exp(1.11999534*math.Log10(float64(fileSize))) + 111248.20149

	if commType == "GET" {
		latency = latency * 0.33
	} else if commType == "PUT" {
		latency = latency * 0.67
	} else {
		log.Panic("Invalid communication type")
	}
	sleepTime := time.Duration(latency) * time.Microsecond
	return sleepTime
}

func (s *server) GetFile(ctx context.Context, in *pb.FileSize) (*pb.FileBlob, error) {
	// Generate random blob
	size := in.GetSize()
	log.Debugf("GET: %d Bytes", size)
	blob := buffer[:size]

	return &pb.FileBlob{Blob: blob}, nil
}

func (s *server) PutFile(ctx context.Context, in *pb.FileBlob) (*pb.FileSize, error) {
	// Wait
	size := int64(len(in.GetBlob()))
	log.Debugf("PUT: %d Bytes", size)

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
	maxMsgSize := int(math.Pow(2, 31)) // 2GB
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

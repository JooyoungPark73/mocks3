package main

import (
	"crypto/rand"
	"flag"
	"io"
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
	buffer    = make([]byte, 2*1024*1024) // 2MB
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

func (s *server) GetFile(req *pb.FileSize, stream pb.FileService_GetFileServer) error {
	size := req.GetSize()

	log.Debugf("GET: %d Bytes", size)

	for remaining := size; remaining > 0; remaining -= int64(len(buffer)) {
		chunk := buffer
		if remaining < int64(len(buffer)) {
			chunk = buffer[:remaining]
		}
		if err := stream.Send(&pb.FileBlob{Blob: chunk}); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func (s *server) PutFile(stream pb.FileService_PutFileServer) error {
	size := int64(0)
	for {
		chunk, err := stream.Recv()
		size += int64(len(chunk.GetBlob()))
		if err == io.EOF {
			log.Debugf("PUT: %d Bytes", size)
			return stream.SendAndClose(&pb.FileSize{Size: size})
		}
		if err != nil {
			return err
		}
	}
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
	// maxMsgSize := int(math.Pow(2, 31) + 1024) // 2GB
	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &server{})
	log.Infof("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

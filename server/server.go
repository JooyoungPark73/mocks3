package main

import (
	"context"
	"crypto/rand"
	"flag"
	"math"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	pb "mocks3/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

var (
	port      = flag.String("port", "50051", "the port to listen on")
	verbosity = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
)

func (s *server) GetTimeToSleep(fileSize int64) time.Duration {
	// Sourced from: https://github.com/vhive-serverless/MockS3/blob/main/mocks3/mock_io_functions.py
	numBytestoKBLog := math.Log10(float64(fileSize) / 1024)
	latencyPower := math.Pow(math.E, 0.0429*numBytestoKBLog) * 2.1114
	sleepTime := time.Duration(math.Pow(10, latencyPower)) * time.Millisecond
	return sleepTime
}

func fillWithRandomData(slice []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	rand.Read(slice) // Fill the slice with random data
}

func (s *server) GetFile(ctx context.Context, in *pb.FileSize) (*pb.FileBlob, error) {
	// Wait based on file size
	arrivalTime := time.Now().UnixMilli()

	// Generate random blob
	blob := make([]byte, in.GetSize())
	if in.GetSize() < 50000000 {
		rand.Read(blob)
	} else {
		const numGoroutines = 10

		var wg sync.WaitGroup
		chunkSize := in.GetSize() / numGoroutines

		for i := 0; i < numGoroutines; i++ {
			start := int64(i) * chunkSize
			end := start + chunkSize
			if i == numGoroutines-1 {
				end = in.GetSize() // Ensure the last chunk covers the rest of the slice
			}

			wg.Add(1)
			go fillWithRandomData(blob[start:end], &wg)
		}
		wg.Wait()
	}
	rand.Read(blob)

	timeToSleep := s.GetTimeToSleep(in.GetSize())
	log.Infof("Received a file blob of size: %d KB, sleeping for %v", in.GetSize()/1024, timeToSleep)
	sleepTime := timeToSleep - time.Duration(time.Now().UnixMilli()-arrivalTime)*time.Millisecond
	log.Info("Net Sleeping for ", sleepTime)
	time.Sleep(sleepTime)

	return &pb.FileBlob{Blob: blob, CreationTime: sleepTime.Milliseconds()}, nil
}

func (s *server) PutFile(ctx context.Context, in *pb.FileBlob) (*pb.FileSize, error) {
	// Wait
	arrivalTime := time.Now().UnixMilli()
	size := int64(len(in.GetBlob()))

	timeToSleep := s.GetTimeToSleep(size)
	log.Infof("Received a file blob of size: %d KB, sleeping for %v", size/1024, timeToSleep)
	sleepTime := timeToSleep - time.Duration(in.GetCreationTime())*time.Millisecond - time.Duration(time.Now().UnixMilli()-arrivalTime)*time.Millisecond
	log.Info("Net Sleeping for ", sleepTime)
	time.Sleep(sleepTime)

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
}

func main() {
	maxMsgSize := int(math.Pow(2, 30)) // 1GB
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

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

	pb "mocks3/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedFileServiceServer
}

var (
	port            = flag.String("port", "30000", "the port to listen on")
	verbosity       = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
	latency         = flag.String("latency", "true", "Whether to have latency, true by default")
	useCreationTime = flag.String("useCreationTime", "true", "Whether to use creation time, true by default")
	buffer          = make([]byte, 512*1024*1024) // 512MB
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
	// Wait based on file size
	arrivalTime := time.Now().UnixMilli()

	// Generate random blob
	size := in.GetSize()
	log.Debugf("Received a request for blob of size: %f KB", float64(size)/1024)
	blob := buffer[:size]
	creationTime := time.Duration(time.Now().UnixMilli()-arrivalTime) * time.Millisecond
	timeToSleep := 0 * time.Millisecond
	sleepTime := 0 * time.Millisecond

	if *latency == "true" {
		timeToSleep = s.GetTimeToSleep("GET", size)
		log.Debugf("Sleeping for %v", timeToSleep)
		if *useCreationTime == "true" {
			creationTime = time.Duration(time.Now().UnixMilli()-arrivalTime) * time.Millisecond
			sleepTime = timeToSleep - creationTime
		} else {
			sleepTime = timeToSleep
		}
		log.Debugf("Creation Time : %v, Net Sleeping for %v", creationTime, sleepTime)
		time.Sleep(sleepTime)
	}

	return &pb.FileBlob{Blob: blob, ExpectedLatency: timeToSleep.Milliseconds(), SleepTime: sleepTime.Milliseconds(), CreationTime: creationTime.Milliseconds()}, nil
}

func (s *server) PutFile(ctx context.Context, in *pb.FileBlob) (*pb.FileSize, error) {
	// Wait
	arrivalTime := time.Now().UnixMilli()
	size := int64(len(in.GetBlob()))
	log.Debugf("Received a file blob of size: %f KB", float64(size)/1024)
	timeToSleep := 0 * time.Millisecond
	sleepTime := 0 * time.Millisecond

	if *latency == "true" {
		timeToSleep = s.GetTimeToSleep("PUT", size)
		log.Debugf("sleeping for %v", timeToSleep)
		if *useCreationTime == "true" {
			sleepTime = timeToSleep - time.Duration(in.GetCreationTime())*time.Millisecond - time.Duration(time.Now().UnixMilli()-arrivalTime)*time.Millisecond
		} else {
			sleepTime = timeToSleep
		}

		log.Debugf("Net Sleeping for %v", sleepTime)
		time.Sleep(sleepTime)
	}

	return &pb.FileSize{Size: size, ExpectedLatency: timeToSleep.Milliseconds(), SleepTime: sleepTime.Milliseconds()}, nil
}

func init() {
	flag.Parse()

	if *latency == "true" {
		log.Info("Latency is enabled")
	} else {
		log.Info("Latency is disabled")
	}

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

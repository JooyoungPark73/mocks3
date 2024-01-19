package mocks3

import (
	"context"
	"flag"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	pb "github.com/JooyoungPark73/mocks3/proto"
	utils "github.com/JooyoungPark73/mocks3/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)

	switch *utils.Verbosity {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func ClientGet(size int64, addr string) (int64, int64) {
	start := time.Now()
	targetTime := utils.GetTimeToSleep("GET", size).Microseconds()

	// get server address from environment variable
	var serverAddress string
	if addr != "none" {
		serverAddress = addr
	} else if _, ok := os.LookupEnv("MOCKS3_SERVER_ADDRESS"); ok {
		serverAddress = os.Getenv("MOCKS3_SERVER_ADDRESS")
	} else {
		serverAddress = *utils.Addr
	}

	// gRPC Connection
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)
	GRPCConnectionEstablishTime := time.Since(start).Microseconds()
	log.Debugf("GRPC Connection Time: %d us", GRPCConnectionEstablishTime)

	// Send the request
	stream, err := c.GetFile(context.Background(), &pb.FileSize{Size: size})
	if err != nil {
		log.Fatalf("client.GetFile Cannot send request size: %v", err)
	}
	recv_size := int64(0)
	for {
		chunk, err := stream.Recv()
		recv_size += int64(len(chunk.GetBlob()))
		// log.Debugf("GET: recvd %d / %d Bytes \r", recv_size, size)
		if err == io.EOF {
			log.Debugf("GET: %d Bytes", recv_size)
			break
		}
		if err != nil {
			log.Fatalf("could not get file from stream: %v", err)
		}
	}
	commTime := time.Since(start).Microseconds() - GRPCConnectionEstablishTime

	log.Debugf("Received blob size: %d Bytes, commTime: %d us", recv_size, commTime)

	timeToSleep := time.Duration(targetTime-time.Since(start).Microseconds()) * time.Microsecond
	time.Sleep(timeToSleep)
	log.Debugf("Time to sleep: %d us, net sleep: %d us", targetTime, timeToSleep.Microseconds())
	e2eTime := time.Since(start).Microseconds()

	return e2eTime, targetTime
}

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

func ClientPut(size int64, addr string) (int64, int64) {
	start := time.Now()
	targetTime := utils.GetTimeToSleep("PUT", size).Microseconds()

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
	stream, err := c.PutFile(context.Background())
	if err != nil {
		log.Fatalf("client.PutFile Connection Failed: %v", err)
	}
	GRPCConnectionEstablishTime := time.Since(start).Microseconds()

	// Generate a random blob
	buffer := utils.CreateRandomObject(2 * 1024 * 1024)
	creationTime := time.Since(start).Microseconds() - GRPCConnectionEstablishTime
	log.Debugf("GRPC Connection Time: %d us, Creation Time: %d us", GRPCConnectionEstablishTime, creationTime)

	// Send the blob
	for remaining := size; remaining > 0; remaining -= int64(len(buffer)) {
		if remaining < int64(len(buffer)) {
			buffer = buffer[:remaining]
		}
		if err := stream.Send(&pb.FileBlob{Blob: buffer}); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("client.PutFile Send Failed: %v", err)
		}
	}

	r, err := stream.CloseAndRecv()
	if err == io.EOF {
		commTime := time.Since(start).Microseconds() - GRPCConnectionEstablishTime - creationTime
		log.Debugf("Sent blob size: %d Bytes, commTime: %d us", r.GetSize(), commTime)
	} else if err != nil {
		log.Fatalf("client.PutFile Recv Failed: %v", err)
	}

	timeToSleep := time.Duration(targetTime-time.Since(start).Microseconds()) * time.Microsecond
	time.Sleep(timeToSleep)
	log.Debugf("Time to sleep: %d us, net sleep: %d us", targetTime, timeToSleep.Microseconds())
	e2eTime := time.Since(start).Microseconds()

	return e2eTime, targetTime
}

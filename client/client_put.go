package mocks3

import (
	"context"
	"flag"
	"math"
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
	maxMsgSize := int(math.Pow(2, 29)) // 512MB
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	GRPCConnectionEstablishTime := time.Since(start).Microseconds()

	// Generate a random blob
	blob := utils.CreateRandomObject(size)
	creationTime := time.Since(start).Microseconds() - GRPCConnectionEstablishTime
	log.Debugf("GRPC Connection Time: %d us, Creation Time: %d us", GRPCConnectionEstablishTime, creationTime)

	// Send the blob
	r, err := c.PutFile(ctx, &pb.FileBlob{Blob: blob})
	commTime := time.Since(start).Microseconds() - GRPCConnectionEstablishTime - creationTime
	if err != nil {
		log.Fatalf("could not get file size: %v", err)
	}
	log.Debugf("Sent blob size: %d Bytes, commTime: %d us", r.GetSize(), commTime)

	timeToSleep := time.Duration(targetTime-time.Since(start).Microseconds()) * time.Microsecond
	time.Sleep(timeToSleep)
	log.Debugf("Time to sleep: %d us, net sleep: %d us", targetTime, timeToSleep.Microseconds())
	e2eTime := time.Since(start).Microseconds()

	return e2eTime, targetTime
}

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

func ClientGet(size int64, addr string) (int64, int64) {
	sendTime := time.Now().UnixMicro()
	targetTime := utils.GetTimeToSleep("GET", size)

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

	// Send the request
	r, err := c.GetFile(ctx, &pb.FileSize{Size: size})
	commTime := time.Now().UnixMicro() - sendTime
	if err != nil {
		if err == context.DeadlineExceeded {
			// handle the timeout
			log.Fatalf("request timed out: %v", err)
		} else {
			log.Fatalf("could not get file: %v", err)
		}
	}
	log.Debugf("Received a file blob of size: %d", len(r.GetBlob()))

	timeToSleep := targetTime - time.Duration(commTime)*time.Microsecond
	time.Sleep(timeToSleep)
	log.Debugf("Time to sleep: %d ms, net sleep: %d ms", targetTime.Milliseconds(), timeToSleep.Milliseconds())
	e2eTime := time.Now().UnixMicro() - sendTime

	return e2eTime / 1000, targetTime.Milliseconds()
}

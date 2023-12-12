package mocks3

import (
	"context"
	"encoding/csv"
	"flag"
	"math"
	"math/rand"
	"os"
	"strconv"
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
	maxMsgSize := int(math.Pow(2, 29)) // 1GB
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Generate a random blob
	blob := utils.CreateRandomObject(size)
	creationTime := time.Now().UnixMicro() - sendTime
	log.Debugf("Creation Time: %d ms", creationTime/1000)

	// Send the blob
	r, err := c.PutFile(ctx, &pb.FileBlob{Blob: blob})
	commTime := time.Now().UnixMicro() - sendTime - creationTime
	if err != nil {
		log.Fatalf("could not get file size: %v", err)
	}
	log.Debugf("Sent a blob, received size: %d", r.GetSize())

	timeToSleep := time.Duration(targetTime.Microseconds()-commTime-creationTime) * time.Microsecond
	time.Sleep(timeToSleep)
	log.Debugf("Time to sleep: %d ms, net sleep: %d ms", targetTime.Milliseconds(), timeToSleep.Milliseconds())
	e2eTime := time.Now().UnixMicro() - sendTime

	return e2eTime / 1000, targetTime.Milliseconds()
}

func BenchmarkClientPut(testIteration int) {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("put_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	err = csvwriter.Write([]string{"Payload Size (Bytes)", "E2E Time (ms)", "Target Time (ms)"})
	if err != nil {
		log.Fatalf("could not write to CSV file: %v", err)
	}

	for i := 0; i < testIteration; i++ {
		randNumber := rand.Float64() * 29
		payloadSize := int64(math.Pow(2, randNumber))
		e2eTime, targetTime := ClientPut(payloadSize)
		err := csvwriter.Write([]string{strconv.FormatInt(payloadSize, 10), strconv.FormatInt(e2eTime, 10), strconv.FormatInt(targetTime, 10)})
		if err != nil {
			log.Fatalf("could not write to CSV file: %v", err)
		}
		csvwriter.Flush()
	}
	csvFile.Close()
}

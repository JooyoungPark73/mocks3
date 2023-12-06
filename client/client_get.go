package main

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

	pb "mocks3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr      = flag.String("addr", "localhost:30000", "the address to connect to")
	verbosity = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
)

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

func ClientGet(size int64) (int64, int64, int64, int64, int64) {
	// Set up a connection to the server.
	sendTime := time.Now().UnixMilli()
	maxMsgSize := int(math.Pow(2, 29)) // 512MB
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r1, err := c.GetFile(ctx, &pb.FileSize{Size: size, ExpectedLatency: 0})
	e2eTime := time.Now().UnixMilli() - sendTime
	if err != nil {
		if err == context.DeadlineExceeded {
			// handle the timeout
			log.Fatalf("request timed out: %v", err)
		} else {
			log.Fatalf("could not get file: %v", err)
		}
	}
	log.Debugf("Received a file blob of size: %d", len(r1.GetBlob()))
	creationTime := r1.GetCreationTime()
	expectedLatency := r1.GetExpectedLatency()
	sleepTime := r1.GetSleepTime()
	commTime := e2eTime - creationTime
	if sleepTime > 0 {
		commTime = commTime - sleepTime
	}
	return e2eTime, expectedLatency, commTime, creationTime, sleepTime
}

func BenchmarkClientGet() {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("get_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	err = csvwriter.Write([]string{"Payload Size (Bytes)", "E2E Time (ms)", "Expected Latency (ms)", "Comm Time (ms)", "Creation Time (ms)", "Sleep Time (ms)"})
	if err != nil {
		log.Fatalf("could not write to CSV file: %v", err)
	}

	for i := 0; i < 1000; i++ {
		randNumber := rand.Float64() * 29
		payloadSize := int64(math.Pow(2, randNumber))
		e2eTime, expectedLatency, commTime, creationTime, sleepTime := ClientGet(payloadSize)
		err := csvwriter.Write([]string{strconv.FormatInt(payloadSize, 10), strconv.FormatInt(e2eTime, 10), strconv.FormatInt(expectedLatency, 10), strconv.FormatInt(commTime, 10), strconv.FormatInt(creationTime, 10), strconv.FormatInt(sleepTime, 10)})
		if err != nil {
			log.Fatalf("could not write to CSV file: %v", err)
		}
		csvwriter.Flush()
	}
	csvFile.Close()
	// Teardown any resources
}

func main() {
	BenchmarkClientGet()
}

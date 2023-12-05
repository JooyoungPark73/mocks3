package main

import (
	"context"
	crand "crypto/rand"
	"encoding/csv"
	"flag"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
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

func fillWithRandomData(slice []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	crand.Read(slice) // Fill the slice with random data
}

func ClientPut(size int64) (int64, int64, int64, int64, int64) {
	// Set up a connection to the server.
	sendTime := time.Now().UnixMilli()
	maxMsgSize := int(math.Pow(2, 29)) // 1GB
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// sendTime := time.Now().UnixMilli()
	blob := make([]byte, size)
	if size < 1024*1024 {
		crand.Read(blob)
	} else {
		buffer := make([]byte, 1024*1024)
		crand.Read(buffer)

		for remaining := size; remaining > 0; remaining -= int64(len(buffer)) {
			if remaining < int64(len(buffer)) {
				buffer = buffer[:remaining]
			}
			copy(blob[size-remaining:], buffer)
		}
	}
	creationTime := time.Now().UnixMilli() - sendTime
	log.Debugf("Creation Time: %d ms", creationTime)
	r2, err := c.PutFile(ctx, &pb.FileBlob{Blob: blob, CreationTime: time.Now().UnixMilli() - sendTime})
	e2eTime := time.Now().UnixMilli() - sendTime
	if err != nil {
		log.Fatalf("could not get file size: %v", err)
	}
	log.Debugf("Sent a blob, received size: %d", r2.GetSize())

	expectedLatency := r2.GetExpectedLatency()
	sleepTime := r2.GetSleepTime()
	commTime := e2eTime - creationTime
	if sleepTime > 0 {
		commTime = commTime - sleepTime
	}
	return e2eTime, expectedLatency, commTime, creationTime, sleepTime
}

func BenchmarkClientPut() {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("put_benchmark.csv")
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
		e2eTime, expectedLatency, commTime, creationTime, sleepTime := ClientPut(payloadSize)
		err := csvwriter.Write([]string{strconv.FormatInt(payloadSize, 10), strconv.FormatInt(e2eTime, 10), strconv.FormatInt(expectedLatency, 10), strconv.FormatInt(commTime, 10), strconv.FormatInt(creationTime, 10), strconv.FormatInt(sleepTime, 10)})
		if err != nil {
			log.Fatalf("could not write to CSV file: %v", err)
		}
		csvwriter.Flush()
	}
	csvFile.Close()
}

func main() {
	BenchmarkClientPut()
}

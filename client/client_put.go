package main

import (
	"context"
	crand "crypto/rand"
	"encoding/csv"
	"flag"
	"math"
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
	addr      = flag.String("addr", "localhost:50051", "the address to connect to")
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

func ClientPut(size int64) int64 {
	// Set up a connection to the server.
	maxMsgSize := int(math.Pow(2, 30)) // 1GB
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFileServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	sendTime := time.Now().UnixMilli()
	blob := make([]byte, size)
	if size < 50000000 {
		crand.Read(blob)
	} else {
		const numGoroutines = 10

		var wg sync.WaitGroup
		chunkSize := size / numGoroutines

		for i := 0; i < numGoroutines; i++ {
			start := int64(i) * chunkSize
			end := start + chunkSize
			if i == numGoroutines-1 {
				end = size // Ensure the last chunk covers the rest of the slice
			}

			wg.Add(1)
			go fillWithRandomData(blob[start:end], &wg)
		}
		wg.Wait()
	}

	log.Info("Creation Time: ", time.Now().UnixMilli()-sendTime)
	r2, err := c.PutFile(ctx, &pb.FileBlob{Blob: blob, CreationTime: time.Now().UnixMilli() - sendTime})
	e2eTime := time.Now().UnixMilli() - sendTime
	if err != nil {
		log.Fatalf("could not get file size: %v", err)
	}
	log.Debugf("Sent a blob, received size: %d", r2.GetSize())
	return e2eTime
}

func BenchmarkClientPut() {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("put_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	e2eTime := int64(0)
	for i := 1; i < int(math.Pow(2, 30)); i = i * 2 {
		e2eTime = ClientPut(int64(i)) // For example, test with a blob size of 1024 bytes
		err := csvwriter.Write([]string{strconv.FormatInt(int64(i), 10), strconv.FormatInt(e2eTime, 10)})
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

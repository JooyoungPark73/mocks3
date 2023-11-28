package main

import (
	"context"
	"encoding/csv"
	"flag"
	"math"
	"os"
	"strconv"
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

func ClientGet(size int64) int64 {
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
	r1, err := c.GetFile(ctx, &pb.FileSize{Size: size})
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
	return e2eTime
}

func BenchmarkClientGet() {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("get_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	e2eTime := int64(0)
	for i := 1; i < int(math.Pow(2, 30)); i = i * 2 {
		e2eTime = ClientGet(int64(i)) // For example, test with a blob size of 1024 bytes
		err := csvwriter.Write([]string{strconv.FormatInt(int64(i), 10), strconv.FormatInt(e2eTime, 10)})
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

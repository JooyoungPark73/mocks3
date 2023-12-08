package mocks3

import (
	"crypto/rand"
	"flag"
	"log"
	"math"
	"time"
)

var (
	Addr      = flag.String("addr", "localhost:30000", "the address to connect to")
	Verbosity = flag.String("verbosity", "info", "Logging verbosity - choose from [info, debug, trace]")
)

func GetTimeToSleep(commType string, fileSize int64) time.Duration {
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

func CreateRandomObject(size int64) []byte {
	blob := make([]byte, size)
	if size < 1024*1024 {
		rand.Read(blob)
	} else {
		buffer := make([]byte, 1024*1024)
		rand.Read(buffer)
		for remaining := size; remaining > 0; remaining -= int64(len(buffer)) {
			if remaining < int64(len(buffer)) {
				buffer = buffer[:remaining]
			}
			copy(blob[size-remaining:], buffer)
		}
	}
	return blob
}

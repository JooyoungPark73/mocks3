package mocks3

import (
	"encoding/csv"
	"math"
	"math/rand"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func BenchmarkClientGet(testIteration int) {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("get_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	err = csvwriter.Write([]string{"Payload Size (Bytes)", "E2E Time (us)", "Target Time (us)"})
	if err != nil {
		log.Fatalf("could not write to CSV file: %v", err)
	}

	for i := 0; i < testIteration; i++ {
		randNumber := rand.Float64() * 29
		payloadSize := int64(math.Pow(2, randNumber))
		e2eTime, targetTime := ClientGet(payloadSize, "none")
		err := csvwriter.Write([]string{strconv.FormatInt(payloadSize, 10), strconv.FormatInt(e2eTime, 10), strconv.FormatInt(targetTime, 10)})
		if err != nil {
			log.Fatalf("could not write to CSV file: %v", err)
		}
		csvwriter.Flush()
	}
	csvFile.Close()
	// Teardown any resources
}

func BenchmarkClientPut(testIteration int) {
	// Setup any required resources (like a mock server)
	csvFile, err := os.Create("put_benchmark.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()
	err = csvwriter.Write([]string{"Payload Size (Bytes)", "E2E Time (us)", "Target Time (us)"})
	if err != nil {
		log.Fatalf("could not write to CSV file: %v", err)
	}

	for i := 0; i < testIteration; i++ {
		randNumber := rand.Float64() * 29
		payloadSize := int64(math.Pow(2, randNumber))
		e2eTime, targetTime := ClientPut(payloadSize, "none")
		err := csvwriter.Write([]string{strconv.FormatInt(payloadSize, 10), strconv.FormatInt(e2eTime, 10), strconv.FormatInt(targetTime, 10)})
		if err != nil {
			log.Fatalf("could not write to CSV file: %v", err)
		}
		csvwriter.Flush()
	}
	csvFile.Close()
}

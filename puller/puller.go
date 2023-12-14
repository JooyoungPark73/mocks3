package main

import (
	"flag"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	mocks3_client "github.com/JooyoungPark73/mocks3/client"
	log "github.com/sirupsen/logrus"
)

var (
	mockS3ServerAddress string
	numberOfWorker      int
	numberOfCPM         int
	imageSize           int
	verbosity           = flag.Lookup("verbosity").Value.(flag.Getter).Get().(string)
)

func getEnvironmentVariables() {
	if _, ok := os.LookupEnv("MOCKS3_SERVER_ADDRESS"); ok {
		mockS3ServerAddress = os.Getenv("MOCKS3_SERVER_ADDRESS")
	} else {
		mockS3ServerAddress = "mocks3-server.default.svc.cluster.local:80"
	}
	log.Infof("MOCKS3_SERVER_ADDRESS = %s", mockS3ServerAddress)

	if _, ok := os.LookupEnv("NUMER_OF_GOWORKER"); ok {
		numberOfWorker, _ = strconv.Atoi(os.Getenv("NUMER_OF_GOWORKER"))
	} else {
		numberOfWorker = 20
	}
	log.Infof("NUMER_OF_GOWORKER = %d", numberOfWorker)

	if _, ok := os.LookupEnv("COLDSTART_PER_MINUTE"); ok {
		numberOfCPM, _ = strconv.Atoi(os.Getenv("COLDSTART_PER_MINUTE"))
	} else {
		numberOfCPM = 60
	}
	log.Infof("COLDSTART_PER_MINUTE = %d", numberOfCPM)

	if _, ok := os.LookupEnv("IMAGE_SIZE"); ok {
		imageSize, _ = strconv.Atoi(os.Getenv("IMAGE_SIZE"))
	} else {
		imageSize = 128
	}
	log.Infof("IMAGE_SIZE = %d MB", imageSize)
}

func init() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)

	getEnvironmentVariables()

	switch verbosity {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func pullImage(cpmPerWorker int) {
	var get_e2e_time int64
	var waitTime time.Duration
	// to avoid all coldstart at the same time
	initialWaitTime := (rand.Float32()*30 + 10)
	log.Infof("Initial wait time: %.2f s", initialWaitTime)
	time.Sleep(time.Duration(initialWaitTime) * time.Second)

	for {
		start := time.Now()
		get_e2e_time, _ = mocks3_client.ClientGet(int64(imageSize*1024*1024), mockS3ServerAddress)

		waitTime = time.Duration(rand.ExpFloat64()*(60/float64(cpmPerWorker))) * time.Second

		timeToSleep := waitTime - time.Since(start)
		log.Infof("Wait: %.2f s, GET: %.2f, net Wait: %.2f s", waitTime.Seconds(), float64(get_e2e_time)/1000000, timeToSleep.Seconds())

		time.Sleep(timeToSleep)
	}
}

func main() {
	cpmPerWorker := numberOfCPM / numberOfWorker
	log.Infof("CPM per worker: %d", cpmPerWorker)

	var wg sync.WaitGroup
	for i := 0; i < numberOfWorker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pullImage(cpmPerWorker)
		}()
	}

	wg.Wait()
}

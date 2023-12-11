package main

import (
	"flag"

	mocks3_client "github.com/JooyoungPark73/mocks3/client"
	mocks3_utils "github.com/JooyoungPark73/mocks3/utils"
)

func init() {
	flag.Parse()
}

func main() {
	mocks3_client.BenchmarkClientPut(*mocks3_utils.TestIteration)
	mocks3_client.BenchmarkClientGet(*mocks3_utils.TestIteration)
}

package main

import (
	mocks3_client "github.com/JooyoungPark73/mocks3/client"
)

func main() {
	mocks3_client.BenchmarkClientPut()
	mocks3_client.BenchmarkClientGet()
}

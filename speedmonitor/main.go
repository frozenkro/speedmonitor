package main

import (
	"fmt"
	"github.com/showwin/speedtest-go/speedtest"
	"log"
)

func main() {
	serverList, _ := speedtest.FetchServers()
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		checkError(s.PingTest(nil))
		checkError(s.DownloadTest())
		checkError(s.UploadTest())

		fmt.Printf("Latency: %s, Download: %s, Upload: %s\n", s.Latency, s.DLSpeed, s.ULSpeed)
		s.Context.Reset()
	}
}
func checkError(err error) {
	if (err != nil) {
		log.Fatal(err)
	}
}

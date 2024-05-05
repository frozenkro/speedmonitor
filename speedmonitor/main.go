package main

import (
	//"context"
	"fmt"
	"log"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	//"github.com/influxdata/influxdb-client-go/v2"
)

type SpeedData struct {
	Latency time.Duration
	DLSpeed speedtest.ByteRate
	ULSpeed speedtest.ByteRate
}

func main() {
	data := runSpeedTest()
	fmt.Printf("Latency: %s, Download: %s, Upload: %s\n", data.Latency, data.DLSpeed, data.ULSpeed)
}

func runSpeedTest() SpeedData {
	serverList, _ := speedtest.FetchServers()
	targets, _ := serverList.FindServer([]int{})

	var data SpeedData

	for _, s := range targets {
		checkError(s.PingTest(nil))
		checkError(s.DownloadTest())
		checkError(s.UploadTest())

		data = SpeedData{s.Latency, s.DLSpeed, s.ULSpeed}
		s.Context.Reset()
	}
	return data
}

func checkError(err error) {
	if (err != nil) {
		log.Fatal(err)
	}
}

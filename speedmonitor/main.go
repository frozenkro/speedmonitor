package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"github.com/joho/godotenv"
	"github.com/go-co-op/gocron"
	"github.com/influxdata/influxdb-client-go/v2"
)

type SpeedData struct {
	Latency time.Duration
	DLSpeed speedtest.ByteRate
	ULSpeed speedtest.ByteRate
	Time time.Time
}

var dbUrl, dbToken, dbOrg, dbBucket string

func init() {
	checkError(godotenv.Load())
}

func main() {

	dbUrl = getEnvVar("SPEEDMONITOR_DB_LOCATION")
	dbToken = getEnvVar("SPEEDMONITOR_DB_TOKEN")
	dbOrg = getEnvVar("SPEEDMONITOR_DB_ORG")
	dbBucket = getEnvVar("SPEEDMONITOR_DB_BUCKET")

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Minute().Do(func() {
		data := runSpeedTest()
		fmt.Printf("Latency: %s, Download: %s, Upload: %s\n", data.Latency, data.DLSpeed, data.ULSpeed)

		saveTest(data)
	})
}

func runSpeedTest() SpeedData {
	serverList, _ := speedtest.FetchServers()
	targets, _ := serverList.FindServer([]int{})

	var data SpeedData

	for _, s := range targets {
		checkError(s.PingTest(nil))
		checkError(s.DownloadTest())
		checkError(s.UploadTest())

		data = SpeedData{s.Latency, s.DLSpeed, s.ULSpeed, time.Now()}
		s.Context.Reset()
	}
	return data
}

func saveTest(data SpeedData) {

	client := influxdb2.NewClient(dbUrl, dbToken)
	writeAPI := client.WriteAPIBlocking(dbOrg, dbBucket)

	sp := influxdb2.NewPointWithMeasurement("speed").
		AddTag("unit", "byte").
		AddField("download", data.DLSpeed).
		AddField("upload", data.ULSpeed).
		SetTime(data.Time)

	lp := influxdb2.NewPointWithMeasurement("latency").
		AddTag("unit", "ms").
		AddField("delay", data.Latency).
		SetTime(data.Time)
		
	sp_err := writeAPI.WritePoint(context.Background(), sp)

	lp_err := writeAPI.WritePoint(context.Background(), lp)

	if sp_err != nil {
		log.Print(sp_err)
	}
	if lp_err != nil {
		log.Print(lp_err)
	}
	
	client.Close()
}

func getEnvVar(varKey string) string {
	val, exists := os.LookupEnv(varKey)
	if !exists {
		log.Fatal("Env var '%s' missing", varKey)
	}
	return val
}

func checkError(err error) {
	if (err != nil) {
		log.Fatal(err)
	}
}

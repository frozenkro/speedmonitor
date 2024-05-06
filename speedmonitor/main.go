package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
	"github.com/showwin/speedtest-go/speedtest"
	"gopkg.in/natefinch/lumberjack.v2"
)

type SpeedData struct {
	Latency time.Duration
	DLSpeed speedtest.ByteRate
	ULSpeed speedtest.ByteRate
	Time    time.Time
}

var dbUrl, dbToken, dbOrg, dbBucket string

func init() {
	checkError(godotenv.Load())
}

func main() {

	logger := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
	}
	defer logger.Close()
	log.SetOutput(logger)

	dbUrl = getEnvVar("SPEEDMONITOR_DB_LOCATION")
	dbToken = getEnvVar("SPEEDMONITOR_DB_TOKEN")
	dbOrg = getEnvVar("SPEEDMONITOR_DB_ORG")
	dbBucket = getEnvVar("SPEEDMONITOR_DB_BUCKET")

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Minute().Do(func() {
		data := runSpeedTest()
		log.Printf("[%s] Latency: %s, Download: %s, Upload: %s\n", time.Now(), data.Latency, data.DLSpeed, data.ULSpeed)

		saveTest(data)
	})
	s.StartBlocking()
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
		log.Printf("[%s] %s\n", time.Now(), sp_err)
	}
	if lp_err != nil {
		log.Printf("[%s] %s\n", time.Now(), lp_err)
	}

	client.Close()
}

func getEnvVar(varKey string) string {
	val, exists := os.LookupEnv(varKey)
	if !exists {
		log.Fatalf("[%s] Env var %s missing\n", time.Now(), varKey)
	}
	return val
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

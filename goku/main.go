package main

import (
	"fg-test/lib"
	"fg-test/pb"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

var (
	addr      string
	d         int64
	rate      int
	connCount uint
)

func init() {
	flag.Int64Var(&d, "d", 10, "duration of tests")
	flag.IntVar(&rate, "rate", 1000, "req per second")
	flag.StringVar(&addr, "addr", "172.16.10.140:8602", "server address")

	flag.UintVar(&connCount, "c", 10, "worker/tcp connections count")
}

func main() {
	flag.Parse()

	fmt.Printf(`Goku is Running %ds test @ %s
Rate : %d/sec:`, d, addr, rate)

	rate := vegeta.Rate{Freq: rate, Per: time.Second}
	duration := time.Second * time.Duration(d)
	rb := lib.Robot{
		UserID:    3030,
		MessageID: pb.MessageID_IDCSLoginReq,
		//ProtoNum:  498753069,
		ProtoNum:  684686029,
	}
	worker := lib.NewWorker(lib.Workers(uint64(connCount)), lib.MaxWorkers(1000))

	var metrics vegeta.Metrics
	for res := range worker.Run(rb, rate, duration, addr) {
		if res.Error != "" {
			fmt.Println(res.Error)
		}
		metrics.Add(res)
	}

	rand.Float32()
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	reporter.Report(os.Stdout)

}

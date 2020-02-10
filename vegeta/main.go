package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

var (
	addr string
	d int64
	rate int
)

func init(){
	flag.Int64Var(&d, "d", 10, "duration of tests")
	flag.IntVar(&rate, "rate", 10000, "req per second")
	flag.StringVar(&addr, "addr", "http://172.16.30.108:8080/login", "server address")
}

func main() {
	flag.Parse()

	fmt.Printf(`Vegeta is Running %ds test @ %s
Rate : %d/sec:`, d, addr, rate)

	rate := vegeta.Rate{Freq: rate, Per: time.Second}
	duration := time.Second * time.Duration(d)
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "POST",
		URL:    addr,
		//URL:    "http://172.16.30.83:8080/login",
		Body:   []byte(`{
							"cuid": "ding",
							"tk": "20190508",
							"ch": 1234
						}`),
	})
	attacker := vegeta.NewAttacker(vegeta.Timeout(5 * time.Second))

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		//if res.Error != ""{
		//	fmt.Println(res.Error)
		//}
		metrics.Add(res)
	}
	metrics.Close()

	reporter := vegeta.NewTextReporter(&metrics)
	reporter.Report(os.Stdout)

}

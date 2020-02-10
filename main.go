package main

import (
	"fg-test/lib"
	"fg-test/pb"
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (

	//测试地址
	addr string
	//开启的连接数
	connCount uint
	//测试持续时长
	duration int64
	//uid起始值
	uidStart     uint32

	rate int

	//结果统计
	totalCount   uint64
	successCount uint64
	bytesIn       uint64
	bytesOut      uint64
	totalLantency int64
	res = Result{}
)

func init() {
	flag.StringVar(&addr, "addr", "172.16.10.140:8602", "example:172.16.10.140:8602")
	flag.UintVar(&connCount, "c", 10, "connection count")
	flag.Int64Var(&duration, "d", 10, "Duration of test(Seconds)")

	flag.IntVar(&rate, "rate", 1000, "req per second")

	uidStart = 10000
}

type Result struct {
	TotalCount   uint64   `json:"total"`
	SuccessCount uint64   `json:"success"`
	TotalLatency uint64   `json:"total_lantency"`
	BytesOut     uint64   `json:"bytes_out"`
	BytesIn      uint64   `json:"bytes_in"`
	Errors       []string `json:"errors"`

	ReqPerSec         uint64 `json:"req_per_sec"`
	AverageLantency   uint64 `json:"avg_lantency"`
	TransferOutPerSec uint64 `json:"trans_out_per_sec"`
	TransferInPerSec  uint64 `json:"trans_int_per_sec"`
}
func main() {

	flag.Parse()

	fmt.Printf(`
Running %ds test @ %s
  %d goroutines and %d connections:
`, duration, addr, connCount, connCount)

	var i uint
	wg := sync.WaitGroup{}

	stop := make(chan bool)
	for i = 0; i < connCount; i++ {
		wg.Add(1)
		go func(workerNum uint, stop <-chan bool) {

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				panic(err)
			}
			client := lib.NewClient(conn)

			defer conn.Close()
			defer wg.Done()

			uid := atomic.AddUint32(&uidStart, 1)
			uidStr := strconv.FormatUint(uint64(uid), 10)
			loginReq := &pb.CSLoginReq{
				Uid:      uid,
				Account:  "test_" + uidStr,
				ProtoNum: 498753069,
				//ProtoNum: 50159572,
			}

			for {
				select {
				case <-stop:
					//fmt.Println("sub goroutine exit")
					return
				default:

					res.TotalCount = atomic.AddUint64(&totalCount, 1)
					nowTime := time.Now()

					//发送消息
					sendLen, err := client.Send(pb.MessageID_IDCSLoginReq, loginReq)
					if err != nil {
						fmt.Println("send buf err:", err.Error())
						res.Errors = append(res.Errors, err.Error())
						continue
					}

					//接收消息
					revLen, err := client.Receive()
					if err != nil {
						fmt.Println("read buf err:", err.Error())
						res.Errors = append(res.Errors, err.Error())
						continue
					}

					//请求成功
					res.SuccessCount = atomic.AddUint64(&successCount, 1)
					res.BytesOut = atomic.AddUint64(&bytesOut, uint64(sendLen))
					res.BytesIn = atomic.AddUint64(&bytesIn, uint64(revLen))
					res.TotalLatency = uint64(atomic.AddInt64(&totalLantency, time.Since(nowTime).Milliseconds()))

					//fmt.Println("running...")
					time.Sleep(10 * time.Millisecond)

				}
			}
		}(i, stop)
	}

	waitForSomeTime()

	close(stop)
	wg.Wait()

	fmt.Println("Test finished!")
	calcResult(&res)
	fmt.Printf(`Average Latency	%dms
Total Latency	%dms
Requests/sec:   %d
TransferOut/sec:   %dKB
TransferIn/sec:   %dKB`, res.AverageLantency, res.TotalLatency, res.ReqPerSec, res.TransferOutPerSec, res.TransferInPerSec)

}

func calcResult(res *Result) {
	res.ReqPerSec = res.SuccessCount / uint64(duration)
	res.AverageLantency = res.TotalLatency / res.SuccessCount
	res.TransferOutPerSec = (res.BytesOut / 1024) / (res.TotalLatency / 1000)
	res.TransferInPerSec = (res.BytesIn / 1024) / (res.TotalLatency / 1000)
}

func waitForSomeTime() {
	ch := time.After(time.Second * time.Duration(duration))
	<-ch
}

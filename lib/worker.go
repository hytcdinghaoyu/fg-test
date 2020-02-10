package lib

import (
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

// Worker is an attack executor which wraps an http.Client
type Worker struct {
	client     Client
	stopch     chan struct{}
	workers    uint64
	maxWorkers uint64
	seqmu      sync.Mutex
	seq        uint64
	began      time.Time
}

const (
	// DefaultTimeout is the default amount of time an Worker waits for a request
	// before it times out.
	DefaultTimeout = 30 * time.Second
	// DefaultWorkers is the default initial number of workers used to carry an attack.
	DefaultWorkers = 10
	// DefaultMaxWorkers is the default maximum number of workers used to carry an attack.
	DefaultMaxWorkers = math.MaxUint64
)

var (
	// DefaultLocalAddr is the default local IP address an Worker uses.
	DefaultLocalAddr = net.IPAddr{IP: net.IPv4zero}
	// DefaultTLSConfig is the default tls.Config an Worker uses.
	DefaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
)

// NewWorker returns a new Worker with default options which are overridden
// by the optionally provided opts.
func NewWorker(opts ...func(*Worker)) *Worker {
	a := &Worker{
		stopch:     make(chan struct{}),
		workers:    DefaultWorkers,
		maxWorkers: DefaultMaxWorkers,
		began:      time.Now(),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Workers returns a functional option which sets the initial number of workers
// an Worker uses to hit its targets. More workers may be spawned dynamically
// to sustain the requested rate in the face of slow responses and errors.
func Workers(n uint64) func(*Worker) {
	return func(a *Worker) { a.workers = n }
}

// MaxWorkers returns a functional option which sets the maximum number of workers
// an Worker can use to hit its targets.
func MaxWorkers(n uint64) func(*Worker) {
	return func(a *Worker) { a.maxWorkers = n }
}

// Timeout returns a functional option which sets the maximum amount of time
// an Worker will wait for a request to be responded to and completely read.
func Timeout(d time.Duration) func(*Worker) {
	return func(a *Worker) {
		a.client.Timeout = d
	}
}

// Attack reads its Targets from the passed Targeter and attacks them at
// the rate specified by the Pacer. When the duration is zero the attack
// runs until Stop is called. Results are sent to the returned channel as soon
// as they arrive and will have their Attack field set to the given name.
func (a *Worker) Run(rb Robot, p vegeta.Pacer, du time.Duration, name string) <-chan *vegeta.Result {
	var wg sync.WaitGroup

	workers := a.workers
	if workers > a.maxWorkers {
		workers = a.maxWorkers
	}

	results := make(chan *vegeta.Result)
	ticks := make(chan struct{})
	for i := uint64(0); i < workers; i++ {
		wg.Add(1)
		go a.attack(rb, name, &wg, ticks, results)
	}

	go func() {
		defer close(results)
		defer wg.Wait()
		defer close(ticks)

		began, count := time.Now(), uint64(0)
		for {
			elapsed := time.Since(began)
			if du > 0 && elapsed > du {
				return
			}

			wait, stop := p.Pace(elapsed, count)
			if stop {
				return
			}

			time.Sleep(wait)

			if workers < a.maxWorkers {
				select {
				case ticks <- struct{}{}:
					count++
					continue
				case <-a.stopch:
					return
				default:
					// all workers are blocked. start one more and try again
					workers++
					wg.Add(1)
					//fmt.Println("all workers are blocked. start one more and try again")
					go a.attack(rb, name, &wg, ticks, results)
				}
			}

			select {
			case ticks <- struct{}{}:
				count++
			case <-a.stopch:
				return
			}
		}
	}()

	return results
}

// Stop stops the current attack.
func (a *Worker) Stop() {
	select {
	case <-a.stopch:
		return
	default:
		close(a.stopch)
	}
}

func (a *Worker) attack(rb Robot, name string, workers *sync.WaitGroup, ticks <-chan struct{}, results chan<- *vegeta.Result) {
	conn, err := net.Dial("tcp", name)
	if err != nil {
		panic(err)
	}
	client := NewClient(conn)

	defer client.Conn.Close()
	defer workers.Done()
	for range ticks {
		results <- a.hit(client, rb, name)
	}
}

func (a *Worker) hit(c Client, rb Robot, name string) *vegeta.Result {
	var (
		res = vegeta.Result{Attack: name}
		err error
	)

	a.seqmu.Lock()
	res.Timestamp = a.began.Add(time.Since(a.began))
	res.Seq = a.seq
	a.seq++
	a.seqmu.Unlock()

	defer func() {
		res.Latency = time.Since(res.Timestamp)
		if err != nil {
			res.Error = err.Error()
		}
	}()

	err, req := rb.Request()
	if err != nil {
		return &res
	}

	//发消息
	sendLen, err := c.Send(rb.MessageID, req)
	if err != nil {
		res.Error = err.Error()
		return &res
	}

	//接收消息
	revLen, err := c.Receive()
	if err != nil {
		fmt.Println(err.Error())
		res.Error = err.Error()
	}

	res.BytesIn = uint64(revLen)
	res.BytesOut = uint64(sendLen)

	if !(res.BytesIn == 0 && res.BytesOut == 0) {
		res.Code = 1
	}

	return &res
}

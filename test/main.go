package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var ticks = make(chan struct{})
	var stopch = make(chan bool)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go attack(&wg, ticks)
	}

	tickCount := 0
	go func() {
		defer fmt.Println("goroutines exited")
		defer wg.Wait()
		defer fmt.Println("waiting for goroutine exit...")
		defer close(ticks)
		for {
			time.Sleep(time.Millisecond * 1)

			select {
			case ticks <- struct{}{}:
				tickCount ++
				fmt.Println("make ticks")
			case <-stopch:
				return

			}
		}
	}()

	time.Sleep(time.Second * 5)
	fmt.Println("close ch")

	fmt.Println(tickCount)
	close(stopch)

}

func attack (wg *sync.WaitGroup, ticks <-chan struct{}){
	defer wg.Done()
	defer fmt.Println("done")
	defer fmt.Println("exit working goroutine..")
	for range ticks{
		//time.Sleep(time.Second)
		fmt.Println("consume ticks...")
	}
}
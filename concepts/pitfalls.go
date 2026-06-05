// CONCEPT 9: Race Conditions, Deadlocks, and Goroutine Leaks
//
// Race condition:
//   Two goroutines access the same memory at the same time, at least one writes,
//   and there is no synchronization. Use `go test -race` or `go run -race`.
//
// Deadlock:
//   Goroutines wait forever, usually on a channel send/receive or a lock.
//
// Goroutine leak:
//   A goroutine stays blocked forever after the caller has moved on.

package concepts

import (
	"fmt"
	"sync"
	"time"
)

func RunPitfallsDemo() {
	fmt.Println("Race demo is intentionally opt-in: go run -race main.go -race-demo")

	unblockWithBuffer()
	unblockWithTimeout()
	preventLeakWithDone()
}

// RunRaceDemo intentionally contains a data race so the race detector can show it.
// Run:
//
//	go run -race main.go -race-demo
func RunRaceDemo() {
	var wg sync.WaitGroup
	counter := 0

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				counter++ // Data race: read-modify-write without a lock.
			}
		}()
	}

	wg.Wait()
	fmt.Printf("Unsafe counter value: %d (race detector should complain)\n", counter)
}

func unblockWithBuffer() {
	// This would deadlock if the channel were unbuffered and no goroutine
	// received immediately. A buffer of 1 lets one send complete.
	ch := make(chan string, 1)
	ch <- "buffered send avoids immediate deadlock"
	fmt.Println(<-ch)
}

func unblockWithTimeout() {
	ch := make(chan string)

	select {
	case msg := <-ch:
		fmt.Println(msg)
	case <-time.After(5 * time.Millisecond):
		fmt.Println("timeout avoided waiting forever on an empty channel")
	}
}

func preventLeakWithDone() {
	done := make(chan struct{})
	values := make(chan int)

	go func() {
		defer close(values)
		for i := 1; ; i++ {
			select {
			case values <- i:
			case <-done:
				fmt.Println("producer exited instead of leaking")
				return
			}
		}
	}()

	fmt.Printf("read first value=%d\n", <-values)
	close(done)
	time.Sleep(5 * time.Millisecond)
}

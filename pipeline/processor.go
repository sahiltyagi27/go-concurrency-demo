// CONCEPT 2: Worker Pool + Buffered Channels
//
// A buffered channel (make(chan T, N)) holds up to N items without blocking.
// Send only blocks when the buffer is FULL. Receive only blocks when EMPTY.
// Use buffered channels to decouple producers from consumers in speed.
//
// Worker Pool: spawn N goroutines that all read from the SAME input channel.
// Go's channel receive is safe across goroutines — only one worker gets each item.
// This gives you bounded concurrency: N workers, no more, no matter the input size.
//
// Pattern:
//   input channel → [worker 1]  ↘
//                  [worker 2]  → results channel
//                  [worker 3]  ↗

package pipeline

import (
	"sync"
	"time"
)

// Process runs a worker pool of `numWorkers` goroutines.
// Each worker reads articles from `in`, processes them, and sends to the output.
//
// sync.WaitGroup coordinates shutdown:
//   - wg.Add(n) before launching goroutines
//   - wg.Done() in each goroutine when it exits
//   - wg.Wait() blocks until all goroutines call Done
//
// Go 1.25+ also supports wg.Go(func() { ... }), which combines Add,
// launching the goroutine, and Done. This file keeps the classic Add/Done
// style because it is still common in existing Go code and makes the mechanics
// explicit for learning.
//
// The closer goroutine waits for all workers to finish, then closes `out`.
// This is the correct shutdown pattern for fan-out → fan-in.
func Process(in <-chan Article, numWorkers int) <-chan Article {
	// Buffered: workers can deposit results without waiting for the aggregator.
	// Buffer size = numWorkers so each worker can deposit one result concurrently.
	out := make(chan Article, numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		workerID := i
		go func() {
			defer wg.Done()
			for article := range in { // range exits when `in` is closed
				// Simulate CPU work (scoring/ranking logic)
				time.Sleep(5 * time.Millisecond)
				article.Score = scoreArticle(article, workerID)
				out <- article
			}
		}()
	}

	// A separate goroutine waits for all workers, then closes the output.
	// We can't close `out` inside the workers — multiple goroutines closing
	// the same channel panics. One closer, one channel.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func scoreArticle(a Article, workerID int) int {
	// Fake scoring: penalize low-score articles, boost by source length
	if a.Score < 20 {
		return a.Score
	}
	return a.Score + len(a.Source) + workerID
}

// CONCEPT 5: Fan-In (Merge)
//
// Fan-in: merge MULTIPLE input channels into ONE output channel.
// Use case: results from N parallel workers combined into one stream.
//
// Each input channel gets its own goroutine forwarding values to the merged output.
// WaitGroup tracks when all forwarders are done so we can close the output.
//
// This is the inverse of fan-out and the natural complement to a worker pool.

package patterns

import "sync"

// FanIn merges multiple input channels into a single output channel.
// Values arrive in whichever order the senders produce them (non-deterministic).
func FanIn[T any](inputs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, in := range inputs {
		in := in // capture loop variable (important pre-Go 1.22)
		go func() {
			defer wg.Done()
			for val := range in {
				out <- val
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

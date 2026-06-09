// CONCEPT 4: Fan-Out
//
// Fan-out: take ONE input channel, duplicate its output to MULTIPLE channels.
// Use case: different consumers need the same data (e.g., logger + processor).
//
// Important: you must read from ALL output channels or the sending goroutine blocks.
// Each output channel must be consumed concurrently.

package patterns

// FanOut reads from `in` and broadcasts each value to all output channels.
// Returns N independent channels, each receiving every value from `in`.
func FanOut[T any](in <-chan T, n int) []<-chan T {
	outputs := make([]chan T, n)
	for i := range outputs {
		outputs[i] = make(chan T, 1)
	}

	go func() {
		defer func() {
			for _, ch := range outputs {
				close(ch)
			}
		}()

		for val := range in {
			// Send to every output channel synchronously.
			// All N consumers receive the same value.
			for _, ch := range outputs {
				ch <- val
			}
		}
	}()

	// Keep bidirectional channels internally so FanOut can send and close them,
	// but return receive-only channels to enforce ownership and prevent callers
	// from sending or closing.
	result := make([]<-chan T, n)
	for i, ch := range outputs {
		result[i] = ch
	}
	return result
}

// CONCEPT 6: select — Multiplexing Channels
//
// select is Go's tool for handling multiple channel operations at once.
// It's like a non-deterministic switch: whichever case is ready runs.
// If multiple cases are ready simultaneously, Go picks one at random.
//
// Key patterns with select:
//   1. Timeout:      case <-time.After(d)
//   2. Non-blocking: default case (runs immediately if no channel is ready)
//   3. Cancellation: case <-done (done channel or context.Done())
//   4. Heartbeat:    case <-ticker.C (periodic work)

package patterns

import (
	"fmt"
	"time"
)

// TimeoutExample shows how to add a deadline to any channel receive.
// If the value arrives within `timeout`, it's returned.
// If not, the zero value and false are returned.
func TimeoutExample[T any](ch <-chan T, timeout time.Duration) (T, bool) {
	select {
	case val := <-ch:
		return val, true
	case <-time.After(timeout):
		var zero T
		return zero, false
	}
}

// NonBlockingReceive tries to receive without waiting.
// Returns immediately whether or not a value is available.
func NonBlockingReceive[T any](ch <-chan T) (T, bool) {
	select {
	case val := <-ch:
		return val, true
	default: // runs instantly if ch has nothing
		var zero T
		return zero, false
	}
}

// HeartbeatWorker demonstrates the ticker + done pattern.
// It does work on every tick and stops when done is closed.
// This is the foundation of background workers, health checkers, etc.
func HeartbeatWorker(interval time.Duration, done <-chan struct{}, work func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			work()
		case <-done:
			fmt.Println("worker: received done signal, shutting down")
			return
		}
	}
}

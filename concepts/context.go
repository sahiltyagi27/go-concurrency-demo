// CONCEPT 7: context.Context
//
// context.Context carries cancellation, deadlines, and request-scoped values.
// In production Go, context is preferred over custom done channels at API
// boundaries because it composes naturally across services, databases, HTTP
// handlers, and goroutines.
//
// Interview line:
//   Use context to tell goroutines "stop working; the caller no longer needs
//   this result" and to enforce deadlines/timeouts.

package concepts

import (
	"context"
	"fmt"
	"time"
)

// RunContextDemo starts two operations:
//   - one cancelled manually by calling cancel()
//   - one stopped automatically by a timeout deadline
func RunContextDemo() {
	manualCtx, cancel := context.WithCancel(context.Background())
	manualResult := fetchProfile(manualCtx, "manual-cancel")

	time.Sleep(20 * time.Millisecond)
	cancel() // Broadcast cancellation to every goroutine using manualCtx.
	fmt.Println(<-manualResult)

	timeoutCtx, stopTimer := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer stopTimer() // Always release timer resources when using WithTimeout.

	timeoutResult := fetchProfile(timeoutCtx, "timeout")
	fmt.Println(<-timeoutResult)
}

func fetchProfile(ctx context.Context, name string) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer close(out)

		select {
		case <-time.After(80 * time.Millisecond):
			out <- fmt.Sprintf("%s: profile loaded", name)
		case <-ctx.Done():
			// ctx.Err() tells you whether it was cancelled or deadline exceeded.
			out <- fmt.Sprintf("%s: stopped early (%v)", name, ctx.Err())
		}
	}()

	return out
}

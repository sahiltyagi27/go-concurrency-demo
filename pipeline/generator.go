// CONCEPT 1: Goroutines + Unbuffered Channels
//
// A goroutine is a lightweight thread managed by the Go runtime (not the OS).
// You spawn one with the "go" keyword. Thousands can run concurrently.
//
// An unbuffered channel (make(chan T)) has NO storage.
// Send BLOCKS until a receiver is ready. Receive BLOCKS until a sender sends.
// This is called a "rendezvous" — both sides must meet.
//
// Pattern: Generator — a goroutine that produces values and owns the channel.
// It creates the channel, launches itself, and returns the read-only channel.
// The caller only receives; only the generator sends and closes.

package pipeline

import (
	"fmt"
	"time"
)

// Article is our data unit flowing through the pipeline.
type Article struct {
	ID     int
	Source string
	Title  string
	Score  int
}

// Generate simulates fetching articles from multiple news sources concurrently.
// It returns a single receive-only channel that emits all articles.
//
// Key ideas:
//   - <-chan Article means "read-only channel" — callers can't accidentally send
//   - The goroutine closes(out) when done, which signals "no more values"
//   - Range over a channel reads until it's closed
func Generate(sources []string) <-chan Article {
	out := make(chan Article) // unbuffered: each send waits for a receiver

	go func() {
		defer close(out) // closing is the signal that the stream is done

		id := 0
		for _, source := range sources {
			// Simulate network latency for each source
			time.Sleep(10 * time.Millisecond)

			// Each source produces 3 articles
			for i := 1; i <= 3; i++ {
				id++
				out <- Article{ // BLOCKS here until someone receives
					ID:     id,
					Source: source,
					Title:  fmt.Sprintf("%s: Article #%d", source, i),
					Score:  i * 10,
				}
			}
		}
	}()

	return out // caller gets the channel BEFORE any articles are sent
}

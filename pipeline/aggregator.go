// CONCEPT 3: Collecting Results + Done Channel (Cancellation)
//
// Done channel pattern: pass a read-only `done <-chan struct{}` to goroutines.
// Closing `done` broadcasts a cancellation signal to ALL goroutines watching it.
// struct{} costs zero bytes — it's the conventional type for signal-only channels.
//
// select statement: like a switch, but for channel operations.
// It picks whichever case is ready. If multiple are ready, it picks randomly.
// This lets a goroutine listen on multiple channels simultaneously.
//
//   select {
//   case v := <-data:   // got a value
//   case <-done:        // cancelled
//   case <-time.After:  // timeout
//   }

package pipeline

import "sort"

// Result is what the aggregator returns after collecting everything.
type Result struct {
	Articles []Article
	Total    int
	Dropped  int // articles dropped due to cancellation
}

// Aggregate drains the `in` channel into a slice.
// It respects `done` — if done is closed, it stops early.
// Returns the final Result with sorted articles (highest score first).
func Aggregate(in <-chan Article, done <-chan struct{}) Result {
	var articles []Article
	dropped := 0

	for {
		select {
		case article, ok := <-in:
			if !ok {
				// Channel closed = pipeline finished normally
				return buildResult(articles, dropped)
			}
			articles = append(articles, article)

		case <-done:
			// Cancellation signal received. Drain remaining buffered items
			// so we don't leave goroutines blocked trying to send.
			dropped = drainAndCount(in)
			return buildResult(articles, dropped)
		}
	}
}

func drainAndCount(in <-chan Article) int {
	count := 0
	for range in {
		count++
	}
	return count
}

func buildResult(articles []Article, dropped int) Result {
	// Sort by score descending
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Score > articles[j].Score
	})
	return Result{
		Articles: articles,
		Total:    len(articles),
		Dropped:  dropped,
	}
}

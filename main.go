// go-concurrency-demo: Concurrent News Feed Pipeline
//
// This program demonstrates every major goroutine/channel concept
// by building a real data pipeline:
//
//   [Generator] ──(articles)──▶ [Worker Pool x3] ──(scored)──▶ [Aggregator]
//
// Run:  go run main.go
// Run with timeout demo:  go run main.go -timeout
//
// Concepts covered (each in its own file):
//   pipeline/generator.go  — goroutines, unbuffered channels, channel ownership
//   pipeline/processor.go  — worker pool, buffered channels, sync.WaitGroup
//   pipeline/aggregator.go — done channel, select, cancellation
//   patterns/fanout.go     — broadcasting one channel to many
//   patterns/fanin.go      — merging many channels into one
//   patterns/select_demo.go — timeout, non-blocking receive, heartbeat
//   concepts/context.go    — context cancellation, timeout, request scoping
//   concepts/mutex.go      — shared memory, Mutex, RWMutex
//   concepts/pitfalls.go   — races, deadlock avoidance, goroutine leaks

package main

import (
	"fmt"
	"os"
	"time"

	"go-concurrency-demo/concepts"
	"go-concurrency-demo/patterns"
	"go-concurrency-demo/pipeline"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-timeout" {
		runWithTimeout()
	} else if len(os.Args) > 1 && os.Args[1] == "-race-demo" {
		concepts.RunRaceDemo()
	} else {
		runFullPipeline()
		runFanOutFanInDemo()
		runSelectDemo()
		runContextDemo()
		runSharedStateDemo()
		runPitfallsDemo()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 1: Full Pipeline
// Shows: generator → worker pool → aggregator, all connected by channels
// ─────────────────────────────────────────────────────────────────────────────

func runFullPipeline() {
	fmt.Println("=== DEMO 1: Full Pipeline ===")

	sources := []string{"TechCrunch", "HackerNews", "ArsTechnica", "TheVerge"}

	// Stage 1: Generate articles (spawns 1 goroutine, returns channel)
	articles := pipeline.Generate(sources)

	// Stage 2: Process with 3 workers (spawns 3 goroutines + 1 closer)
	// Workers read from `articles` concurrently — first free worker gets each item
	scored := pipeline.Process(articles, 3)

	// Stage 3: Aggregate (blocks main goroutine until pipeline drains)
	done := make(chan struct{}) // not cancelled here
	result := pipeline.Aggregate(scored, done)

	fmt.Printf("Collected %d articles, dropped %d\n", result.Total, result.Dropped)
	fmt.Println("Top 5 by score:")
	for i, a := range result.Articles {
		if i >= 5 {
			break
		}
		fmt.Printf("  [%d] %-12s %q  score=%d\n", a.ID, a.Source, a.Title, a.Score)
	}
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 2: Fan-Out → Fan-In
// Shows: one source → 2 independent consumers → merged back
// ─────────────────────────────────────────────────────────────────────────────

func runFanOutFanInDemo() {
	fmt.Println("=== DEMO 2: Fan-Out → Fan-In ===")

	// Create a simple integer source
	source := make(chan int)
	go func() {
		defer close(source)
		for i := 1; i <= 6; i++ {
			source <- i
		}
	}()

	// Fan-out: duplicate source to 2 channels
	// Both channels receive every value from source
	outputs := patterns.FanOut(source, 2)

	// Each consumer transforms the values independently
	doubled := transform(outputs[0], func(n int) int { return n * 2 })
	squared := transform(outputs[1], func(n int) int { return n * n })

	// Fan-in: merge results back into one channel
	merged := patterns.FanIn(doubled, squared)

	fmt.Print("Merged results (order non-deterministic): ")
	for val := range merged {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
	fmt.Println()
}

func transform(in <-chan int, fn func(int) int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- fn(v)
		}
	}()
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 3: select — timeout, non-blocking, heartbeat
// ─────────────────────────────────────────────────────────────────────────────

func runSelectDemo() {
	fmt.Println("=== DEMO 3: select patterns ===")

	// Non-blocking receive
	ch := make(chan string, 1)
	val, ok := patterns.NonBlockingReceive(ch)
	fmt.Printf("Non-blocking (empty channel): val=%q ok=%v\n", val, ok)

	ch <- "hello"
	val, ok = patterns.NonBlockingReceive(ch)
	fmt.Printf("Non-blocking (has value):     val=%q ok=%v\n", val, ok)

	// Timeout: slow channel
	slow := make(chan string)
	go func() {
		time.Sleep(50 * time.Millisecond)
		slow <- "finally arrived"
	}()

	val, ok = patterns.TimeoutExample(slow, 10*time.Millisecond) // too short
	fmt.Printf("Timeout (10ms, slow=50ms):    val=%q ok=%v\n\n", val, ok)

	done := make(chan struct{})
	beats := 0
	go patterns.HeartbeatWorker(10*time.Millisecond, done, func() {
		beats++
		fmt.Printf("Heartbeat tick #%d\n", beats)
	})
	time.Sleep(35 * time.Millisecond)
	close(done)
	time.Sleep(5 * time.Millisecond) // give worker time to print shutdown line
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 4: Cancellation with done channel (-timeout flag)
// Shows: pipeline cut short by closing the done channel
// ─────────────────────────────────────────────────────────────────────────────

func runWithTimeout() {
	fmt.Println("=== DEMO 4: Cancellation via done channel ===")

	sources := []string{"CNN", "BBC", "Reuters", "AP", "NPR", "Bloomberg"}
	articles := pipeline.Generate(sources)
	scored := pipeline.Process(articles, 2)

	done := make(chan struct{})

	// Cancel after 30ms — pipeline won't finish all sources in time
	go func() {
		time.Sleep(30 * time.Millisecond)
		fmt.Println("main: closing done channel (cancellation broadcast)")
		close(done) // closing broadcasts to ALL readers of this channel
	}()

	result := pipeline.Aggregate(scored, done)
	fmt.Printf("Collected %d articles before cancellation, dropped %d buffered\n",
		result.Total, result.Dropped)
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 5: context.Context
// Shows: cancellation, deadlines, and passing request lifetime across goroutines
// ─────────────────────────────────────────────────────────────────────────────

func runContextDemo() {
	fmt.Println("=== DEMO 5: context cancellation and timeout ===")
	concepts.RunContextDemo()
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 6: Shared State
// Shows: Mutex, RWMutex, and why plain shared variables race
// ─────────────────────────────────────────────────────────────────────────────

func runSharedStateDemo() {
	fmt.Println("=== DEMO 6: shared state with Mutex / RWMutex ===")
	concepts.RunSharedStateDemo()
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo 7: Concurrency Pitfalls
// Shows: races, deadlock avoidance, goroutine leaks
// ─────────────────────────────────────────────────────────────────────────────

func runPitfallsDemo() {
	fmt.Println("=== DEMO 7: race, deadlock, and leak prevention ===")
	concepts.RunPitfallsDemo()
	fmt.Println()
}

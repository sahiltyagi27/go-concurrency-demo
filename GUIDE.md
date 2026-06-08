# Go Goroutines & Channels — Deep Dive Guide

> Reference for the `go-concurrency-demo` project.
> Run demos: `go run .`, `go run . -timeout`, or `go run -race . -race-demo`

---

## The Core Philosophy

> "Don't communicate by sharing memory; share memory by communicating."

In most languages you use shared variables + locks for concurrency. In Go, goroutines talk through channels — no shared state, no race conditions.

---

## 1. Goroutine

A goroutine is a function running concurrently with everything else. Launched with the `go` keyword. Goroutines are cheap — you can run millions. The Go runtime schedules them across OS threads automatically.

```go
// Normal function call — blocks until done
doWork()

// Goroutine — returns immediately, doWork runs in background
go doWork()

// Anonymous goroutine — most common pattern
go func() {
    fmt.Println("running concurrently")
}()
```

**In this project:** `pipeline/generator.go` — the generator launches itself as a goroutine and returns the channel before any articles are sent.

```go
func Generate(sources []string) <-chan Article {
    out := make(chan Article)

    go func() {           // ← spawns goroutine
        defer close(out)  // ← signals "no more values" when done
        for _, source := range sources {
            out <- Article{...} // sends articles one by one
        }
    }()

    return out // returned BEFORE goroutine produces anything
}
```

---

## 2. Channels

A channel is a typed pipe between goroutines. One goroutine sends, another receives.

```go
ch := make(chan int)  // create a channel of ints

ch <- 42             // SEND — blocks until someone receives
val := <-ch          // RECEIVE — blocks until someone sends
```

### Unbuffered Channel

No storage. Sender and receiver must be ready at the same time — like a handshake.

```go
ch := make(chan int)  // unbuffered

// This blocks forever — nobody is receiving
ch <- 1  // ← DEADLOCK if no goroutine is receiving
```

```go
// Correct: receiver in a goroutine
go func() { fmt.Println(<-ch) }()
ch <- 1  // now this works — receiver is waiting
```

### Buffered Channel

Has a queue of N slots. Sender only blocks when the buffer is **full**. Receiver only blocks when the buffer is **empty**.

```go
ch := make(chan int, 3)  // buffer of 3

ch <- 1  // doesn't block — goes into buffer
ch <- 2  // doesn't block
ch <- 3  // doesn't block
ch <- 4  // BLOCKS — buffer is full, waiting for a receiver
```

**Use buffered channels** when producer and consumer run at different speeds, to avoid unnecessary blocking.

### Channel Direction (Read-only / Write-only)

```go
func both() chan int { ... }          // returns bidirectional channel
func producer() <-chan int { ... }     // returns read-only channel
func inbox() chan<- int { ... }        // returns write-only channel
func consumer(in <-chan int)  { ... }  // accepts read-only channel
func sender(out chan<- int)   { ... }  // accepts write-only channel
```

If there is no arrow, the channel is bidirectional:

```go
ch := make(chan int) // type is chan int

ch <- 10  // send allowed
v := <-ch // receive allowed
```

Directional channels enforce ownership:

```go
chan int   // send and receive
<-chan int // receive-only / read-only
chan<- int // send-only / write-only
```

Write-only return example:

```go
func StartPrinter() chan<- string {
    ch := make(chan string)

    go func() {
        for msg := range ch {
            fmt.Println(msg)
        }
    }()

    return ch // caller can send, but cannot receive
}

func main() {
    printer := StartPrinter()
    printer <- "hello"
    close(printer)
}
```

Returning `<-chan T` is more common because generators usually expose values to callers. Returning `chan<- T` is useful when you want callers to submit work/messages but not read internal results.

### Closing a Channel

```go
close(ch)        // signals "no more values"
val, ok := <-ch  // ok is false when channel is closed and empty
for v := range ch { ... }  // loop exits automatically when channel closes
```

**Rule:** Only the **sender** closes a channel. Closing from the receiver side or closing twice panics.

---

## 3. Worker Pool — `pipeline/processor.go`

Instead of processing items one by one, spawn N goroutines that all read from the same input channel. Go ensures each item goes to exactly one worker.

```
articles channel ──▶ [worker 1] ──▶
                ──▶ [worker 2] ──▶  results channel
                ──▶ [worker 3] ──▶
```

```go
func Process(in <-chan Article, numWorkers int) <-chan Article {
    out := make(chan Article, numWorkers)  // buffered output
    var wg sync.WaitGroup
    wg.Add(numWorkers)

    for i := 0; i < numWorkers; i++ {
        go func() {
            defer wg.Done()
            for article := range in {  // multiple goroutines, same channel — safe
                article.Score = score(article)
                out <- article
            }
        }()
    }

    // Separate goroutine closes output after all workers finish
    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### sync.WaitGroup

Coordinates waiting for a group of goroutines to finish.

```go
var wg sync.WaitGroup

wg.Add(3)   // I'm launching 3 goroutines

go func() { defer wg.Done(); doWork() }()
go func() { defer wg.Done(); doWork() }()
go func() { defer wg.Done(); doWork() }()

wg.Wait()   // blocks here until all 3 call Done()
```

**Rule:** Never close a channel from multiple goroutines — only one designated "closer" should do it. Use WaitGroup to know when that's safe.

### sync.WaitGroup.Go

Go 1.25 added `WaitGroup.Go`, a helper that starts a goroutine and automatically tracks it in the wait group.

Classic pattern:

```go
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
```

Newer pattern:

```go
wg.Go(func() {
    doWork()
})
```

This reduces mistakes like calling `Add` inside the goroutine or forgetting `Done`. The classic `Add` / `Done` pattern is still important because many existing Go codebases use it.

---

## 4. Done Channel — Cancellation — `pipeline/aggregator.go`

How do you stop goroutines early? Pass them a `done` channel. Closing it broadcasts a signal to **all** goroutines watching it simultaneously.

```go
done := make(chan struct{})  // struct{} = zero bytes, signal-only

// In goroutines:
select {
case val := <-work:
    process(val)
case <-done:
    return  // exit when cancelled
}

// To cancel everything at once:
close(done)  // ALL goroutines watching done will wake up
```

**Why `struct{}`?** It uses zero memory. You never send a value on it — only close it.

```go
// In this project — aggregator stops early when cancelled
func Aggregate(in <-chan Article, done <-chan struct{}) Result {
    var articles []Article
    for {
        select {
        case article, ok := <-in:
            if !ok {
                return buildResult(articles)  // pipeline finished normally
            }
            articles = append(articles, article)
        case <-done:
            return buildResult(articles)  // cancelled early
        }
    }
}
```

---

## 5. select — `patterns/select_demo.go`

`select` lets a goroutine wait on multiple channels simultaneously. Whichever is ready first runs. If multiple are ready, Go picks randomly.

```go
select {
case v := <-ch1:          // received from ch1
case v := <-ch2:          // received from ch2
case ch3 <- val:          // sent to ch3
case <-time.After(5 * time.Second):  // timeout
case <-done:              // cancellation
default:                  // runs immediately if nothing else is ready (non-blocking)
}
```

### Timeout Pattern

```go
func withTimeout[T any](ch <-chan T, timeout time.Duration) (T, bool) {
    select {
    case val := <-ch:
        return val, true
    case <-time.After(timeout):
        var zero T
        return zero, false  // timed out
    }
}
```

### Non-Blocking Receive

```go
select {
case val := <-ch:
    fmt.Println("got:", val)
default:
    fmt.Println("nothing ready, moving on")
}
```

### Heartbeat Worker

```go
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        doPeriodicWork()
    case <-done:
        return  // graceful shutdown
    }
}
```

---

## 6. Fan-Out — `patterns/fanout.go`

Take one channel, broadcast its values to N independent channels. Each consumer gets every value.

```
source ──▶ [fan-out] ──▶ channel A (consumer 1)
                    ──▶ channel B (consumer 2)
```

```go
func FanOut[T any](in <-chan T, n int) []<-chan T {
    outputs := make([]chan T, n)
    for i := range outputs {
        outputs[i] = make(chan T, 1)
    }

    go func() {
        defer func() {
            for _, ch := range outputs { close(ch) }
        }()
        for val := range in {
            for _, ch := range outputs {
                ch <- val  // send to every consumer
            }
        }
    }()
    // ...
}
```

**Warning:** You must consume **all** output channels. If one consumer stops reading, the fan-out goroutine blocks.

---

## 7. Fan-In (Merge) — `patterns/fanin.go`

Merge N channels into one. Values arrive in whatever order the senders produce them.

```
channel A ──▶
channel B ──▶ [fan-in] ──▶ merged channel
channel C ──▶
```

```go
func FanIn[T any](inputs ...<-chan T) <-chan T {
    out := make(chan T)
    var wg sync.WaitGroup
    wg.Add(len(inputs))

    for _, in := range inputs {
        in := in  // capture loop variable
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
```

---

## 8. context.Context — `concepts/context.go`

`context.Context` carries cancellation, deadlines, and request-scoped values across goroutines and API boundaries.

Interview answer:

> Context lets a caller tell downstream work to stop, usually because the request was cancelled, timed out, or no longer needs the result.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

select {
case result := <-work:
    return result
case <-ctx.Done():
    return ctx.Err()
}
```

### `done` channel vs `context`

| Pattern | Best For |
|---|---|
| `done chan struct{}` | Small internal pipelines |
| `context.Context` | Public APIs, HTTP handlers, DB calls, RPC calls, request-scoped cancellation |

**Rule:** accept `context.Context` as the first argument when a function does work that may block, call external systems, or run in a goroutine.

```go
func FetchUser(ctx context.Context, id string) (User, error) {
    // check ctx.Done() while waiting for slow work
}
```

---

## 9. Mutex and RWMutex — `concepts/mutex.go`

Go encourages channels, but shared memory is still normal. When multiple goroutines access the same variable and at least one writes, protect it.

### `sync.Mutex`

Use `Mutex` when one goroutine at a time may enter a critical section.

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

### `sync.RWMutex`

Use `RWMutex` when reads are frequent and writes are less frequent. Multiple readers can hold `RLock` together, but writers need exclusive `Lock`.

```go
mu.RLock()
value := cache[key]
mu.RUnlock()

mu.Lock()
cache[key] = value
mu.Unlock()
```

Interview line:

> Channels coordinate ownership and communication; mutexes protect shared memory. Senior Go code uses both, depending on the shape of the problem.

---

## 10. Race Conditions — `concepts/pitfalls.go`

A data race happens when:

1. Two or more goroutines access the same memory.
2. At least one access is a write.
3. There is no synchronization such as a mutex, channel handoff, or atomic operation.

Unsafe:

```go
counter++ // read + add + write; not atomic
```

Safe:

```go
mu.Lock()
counter++
mu.Unlock()
```

Run the intentional race demo:

```bash
go run -race . -race-demo
```

Run normal code with the race detector:

```bash
go test -race ./...
go run -race .
```

The race detector is one of the most important Go debugging tools to mention in interviews.

---

## 11. Deadlocks and Goroutine Leaks — `concepts/pitfalls.go`

### Deadlock

A deadlock means goroutines are waiting forever and no goroutine can make progress.

Common causes:

| Cause | Example |
|---|---|
| Send with no receiver | `ch <- 1` on an unbuffered channel |
| Receive with no sender | `<-ch` when nobody sends |
| Forgetting to close | `for v := range ch` waits forever |
| Lock ordering bug | goroutine A holds lock 1 and wants lock 2; goroutine B holds lock 2 and wants lock 1 |

Avoid waiting forever with `select`:

```go
select {
case v := <-ch:
    use(v)
case <-time.After(time.Second):
    return errors.New("timed out")
}
```

### Goroutine leak

A goroutine leak means a goroutine stays alive after it is no longer useful, usually blocked on send, receive, sleep, or I/O.

Prevent leaks by giving long-running goroutines a cancellation path:

```go
for {
    select {
    case out <- value:
    case <-done:
        return
    }
}
```

---

## How It All Connects — The Pipeline

```
main.go:

sources := []string{"TechCrunch", "HackerNews", ...}

articles := pipeline.Generate(sources)      // 1 goroutine, unbuffered channel
scored   := pipeline.Process(articles, 3)   // 3 goroutines, buffered channel
result   := pipeline.Aggregate(scored, done) // main goroutine, select + done channel
```

Each stage is a goroutine. They communicate only through channels. The data flows left to right, and each stage can run at its own speed.

```
[Generate] ──(unbuffered)──▶ [Worker 1]
                             [Worker 2] ──(buffered)──▶ [Aggregate]
                             [Worker 3]
```

---

## Common Mistakes

| Mistake | What Happens | Fix |
|---|---|---|
| Sending on a closed channel | panic | Only the sender closes |
| Closing a channel twice | panic | One owner, one close |
| Not closing a channel | goroutine leaks (stuck on `range`) | Always `defer close(ch)` in the sender |
| Closing from the wrong goroutine | data race / panic | Use WaitGroup + a single closer goroutine |
| Forgetting to consume a fan-out channel | blocks the fan-out goroutine | Always consume all outputs concurrently |
| Writing shared state without a lock | data race | Use Mutex, channel ownership, or atomic |
| Starting goroutines without cancellation | goroutine leaks | Pass context or done channel |
| Using `time.After` repeatedly in hot loops | extra timer allocations | Prefer `time.NewTimer` or `time.NewTicker` when appropriate |

---

## Quick Reference

```go
// Create
ch := make(chan T)        // unbuffered
ch := make(chan T, n)     // buffered, size n

// Send / Receive
ch <- val                 // send (blocks if full/no receiver)
val := <-ch               // receive (blocks if empty/no sender)
val, ok := <-ch           // ok=false means channel closed

// Close
close(ch)                 // sender closes, signals end of stream

// Range (exits when channel closed)
for val := range ch { ... }

// Select
select {
case v := <-ch: ...
case <-done:    ...
default:        ...       // non-blocking
}

// Goroutine
go func() { ... }()

// WaitGroup
var wg sync.WaitGroup
wg.Add(n)
defer wg.Done()   // inside goroutine
wg.Wait()         // blocks until all Done()

// WaitGroup.Go, Go 1.25+
wg.Go(func() {
    doWork()
})

// Context
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
<-ctx.Done()

// Mutex
mu.Lock()
shared++
mu.Unlock()

// RWMutex
mu.RLock()
value := shared
mu.RUnlock()
```

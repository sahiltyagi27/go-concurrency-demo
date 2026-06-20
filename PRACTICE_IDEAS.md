# Go Concurrency Practice Ideas

Use these drills to move from knowing concepts to writing concurrency code under interview pressure.

## Beginner

### 1. Print Numbers With Goroutines

- Start 3 goroutines.
- Each goroutine prints a different range.
- Use `sync.WaitGroup` so `main` waits.

Concepts:

```text
goroutine
sync.WaitGroup
wg.Add / wg.Done / wg.Wait
```

### 2. Producer Consumer

- One goroutine sends numbers into a channel.
- Another goroutine receives numbers and prints squares.

Concepts:

```text
channel send
channel receive
close channel
for range over channel
```

### 3. Buffered Channel Demo

- Create `make(chan int, 3)`.
- Send 3 values without a receiver.
- Observe that the 4th send blocks until someone receives.

Concepts:

```text
buffered channel
capacity
blocking send
blocking receive
```

### 4. Timeout With Select

- Start a slow goroutine.
- Use `select` with `time.After`.

Concepts:

```text
select
time.After
timeout
```

## Intermediate

### 5. Worker Pool

- Create `jobs chan Job`.
- Start a fixed number of workers.
- Send jobs.
- Collect results.
- Close channels properly.

Concepts:

```text
bounded concurrency
jobs channel
results channel
WaitGroup
close results after workers finish
```

### 6. Fan-Out / Fan-In

- One input channel.
- Multiple workers process values.
- Merge all worker outputs into one result channel.

Concepts:

```text
fan-out
fan-in
merge channels
WaitGroup
```

### 7. Context Cancellation

- Start a long-running worker.
- Stop it using `context.WithCancel`.
- Also try `context.WithTimeout`.

Concepts:

```text
context.Context
ctx.Done()
cancel function
timeout
```

### 8. Rate Limiter

- Use `time.Ticker`.
- Allow only one request every 200ms.
- Later try burst support with a buffered channel.

Concepts:

```text
time.Ticker
throttling
burst capacity
```

### 9. Pipeline

- Stage 1: generate numbers.
- Stage 2: square numbers.
- Stage 3: filter even results.
- Stage 4: collect output.

Concepts:

```text
pipeline stages
channel ownership
close output channel
compose stages
```

## Advanced / Interview Useful

### 10. Graceful Shutdown

- HTTP server plus background worker.
- On `Ctrl+C`:
  - stop accepting requests
  - cancel workers
  - wait for cleanup

Concepts:

```text
signal.NotifyContext
http.Server.Shutdown
context cancellation
WaitGroup
shutdown timeout
```

### 11. Concurrent Web Scraper

- Given a list of URLs.
- Fetch with max 5 concurrent workers.
- Use context timeout.
- Collect status codes.

Concepts:

```text
worker pool
HTTP client timeout
context
result collection
```

### 12. Retry Worker With Backoff

- Jobs sometimes fail.
- Retry failed jobs up to 3 times.
- Use exponential backoff.
- Send final failures to a DLQ channel.

Concepts:

```text
retry
exponential backoff
max attempts
DLQ channel
```

### 13. Safe Counter

- Increment a shared counter from 100 goroutines.
- First see race with `go run -race`.
- Fix with `sync.Mutex`.
- Try `atomic.Int64`.

Concepts:

```text
race condition
sync.Mutex
atomic.Int64
race detector
```

### 14. Pub/Sub In Memory

- Multiple subscribers.
- Publisher broadcasts messages.
- Subscribers receive independently.
- Handle slow subscribers.

Concepts:

```text
broadcast
subscriber channels
slow consumer handling
non-blocking send
mutex-protected subscriber list
```

### 15. Batch Processor

- Read events from a channel.
- Flush batch when:
  - batch size reaches 10
  - or 2 seconds pass

Concepts:

```text
batching
time.Ticker
select
flush by size
flush by time
```

## Best Practice Order

```text
1. Producer consumer
2. Worker pool
3. Context cancellation
4. Graceful shutdown
5. Rate limiter
6. Retry worker with backoff
7. Batch processor
8. Concurrent web scraper
```

## Interview Goal

You should be able to write these from memory:

```text
worker pool
context cancellation
graceful shutdown
rate limiter
retry with backoff
batch processor
```

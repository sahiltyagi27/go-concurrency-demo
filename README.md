# Go Concurrency Demo

Runnable Go examples for learning concurrency patterns used in backend interviews.

## Run

```bash
go run .
go run . -timeout
go run -race . -race-demo
```

## Topic Index

| Topic | File |
|---|---|
| Goroutines and unbuffered channels | [pipeline/generator.go](pipeline/generator.go) |
| Worker pool, buffered channels, WaitGroup | [pipeline/processor.go](pipeline/processor.go) |
| Done channel, select, cancellation | [pipeline/aggregator.go](pipeline/aggregator.go) |
| Fan-out | [patterns/fanout.go](patterns/fanout.go) |
| Fan-in | [patterns/fanin.go](patterns/fanin.go) |
| Select, timeout, non-blocking receive, heartbeat | [patterns/select_demo.go](patterns/select_demo.go) |
| Context cancellation and deadlines | [concepts/context.go](concepts/context.go) |
| Mutex and RWMutex | [concepts/mutex.go](concepts/mutex.go) |
| Race conditions, deadlocks, goroutine leaks | [concepts/pitfalls.go](concepts/pitfalls.go) |

## Main Guide

Read the full study notes here:

- [GUIDE.md](GUIDE.md)

## Interview Highlights

- Goroutines are lightweight concurrent functions.
- Channels communicate values between goroutines.
- Worker pools provide bounded concurrency.
- `select` coordinates multiple channel operations.
- `context.Context` carries cancellation and deadlines.
- Use mutexes for shared memory and channels for communication.
- Use the race detector with `go run -race` or `go test -race`.


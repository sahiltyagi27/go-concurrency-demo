# Go Concurrency Decision Guide

This note answers the real interview question behind Go concurrency:

> I know goroutines, channels, WaitGroups, worker pools, graceful shutdown, and select. How do I decide which one to use in a real system?

The short answer is:

> Choose the tool based on ownership, coordination, cancellation, and limits.

## Mental Model

```text
goroutine   = run this work concurrently
channel     = pass ownership of data/events between goroutines
WaitGroup   = wait for a known group of goroutines to finish
context     = cancel or timeout work
mutex       = protect shared state
worker pool = limit concurrency for many similar jobs
select      = wait on multiple channel events
```

## Goroutine Without Channel

Use a goroutine without a channel when you need to run background work and you do not need to immediately collect a result.

Examples:

```go
go sendEmail(user)
go writeAuditLog(event)
go publishMetric(metric)
```

This is useful for fire-and-forget style work, but be careful. If the goroutine fails, the caller may never know. In production, important background work is often pushed to a queue, job table, or worker system instead of being launched blindly.

Interview wording:

> I use a goroutine by itself when the work can safely run independently and I do not need the result in the current flow. If failure matters, I avoid pure fire-and-forget and use a queue or return an error through a channel/errgroup.

## Channel

Use a channel when goroutines need to communicate or transfer ownership of work/results.

Common cases:

```text
send jobs to workers
collect results from workers
stream events
fan-in and fan-out
signal cancellation with a done channel
```

Example:

```go
jobs := make(chan Job)
results := make(chan Result)
```

Channels are best when the design is naturally about passing values from one stage to another.

Interview wording:

> I use channels when goroutines need to exchange data, stream values, or coordinate ownership. A channel is not just a queue; it is also a synchronization point.

## WaitGroup

Use `sync.WaitGroup` when you started a known number of goroutines and must wait for all of them to finish.

Example:

```go
var wg sync.WaitGroup

for _, user := range users {
    wg.Add(1)
    go func(user User) {
        defer wg.Done()
        sendEmail(user)
    }(user)
}

wg.Wait()
```

Important rule:

```text
Call wg.Add before starting the goroutine.
Call defer wg.Done inside the goroutine.
Call wg.Wait from the goroutine that needs to wait.
```

Interview wording:

> I use WaitGroup only for waiting. It does not pass data, cancel work, or collect errors.

## Where Not To Use WaitGroup

Do not use `WaitGroup` to pass data.

Bad mental model:

```text
WaitGroup is not communication.
WaitGroup is not cancellation.
WaitGroup is not error handling.
```

If you need a result, use:

```text
channel
errgroup
shared variable protected by mutex
```

Do not use `WaitGroup` as the main control mechanism for goroutines that live forever, such as:

```text
HTTP server
Kafka consumer
background scheduler
long-running worker process
```

For long-running services, use `context`, signal handling, server shutdown, and lifecycle management.

## Context

Use `context.Context` when work is tied to a request lifetime, timeout, cancellation, or graceful shutdown.

Example:

```go
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultCh:
    return result
}
```

Common cases:

```text
HTTP request cancelled by client
database query timeout
external API timeout
service shutdown
worker should stop accepting new jobs
```

Interview wording:

> I pass context through the call chain so cancellation and deadlines are respected. I do not store context in structs for normal request handling; I pass it as the first argument.

## Worker Pool

Use a worker pool when there are many tasks but you want only limited concurrency.

Example scenario:

```text
10,000 jobs
20 workers
external API can handle only limited parallel requests
```

Worker pools protect:

```text
database
Redis
external APIs
CPU
memory
downstream services
```

Interview wording:

> If the number of tasks is large or the downstream system has limits, I use a worker pool. It gives bounded concurrency instead of starting unlimited goroutines.

## Select

Use `select` when a goroutine needs to wait on multiple channel operations.

Example:

```go
select {
case job := <-jobs:
    process(job)
case <-ctx.Done():
    return
case <-time.After(2 * time.Second):
    return errors.New("timeout")
}
```

Common cases:

```text
wait for work or cancellation
wait for result or timeout
non-blocking send/receive with default
heartbeat loops
graceful shutdown loops
```

Interview wording:

> `select` is like a switch for channel operations. It lets one goroutine react to work, cancellation, timeout, or default behavior.

## Anonymous Function Vs Named Function Goroutine

You can start a goroutine with either a direct function call or an anonymous function.

Direct function call:

```go
go sendEmail(user)
go worker(ctx, jobs, results)
go server.ListenAndServe()
```

Anonymous function:

```go
go func() {
    defer wg.Done()
    process(job)
}()
```

Use an anonymous function when you need small wrapper logic around the goroutine.

Common cases:

```text
defer wg.Done()
recover from panic
select on ctx.Done()
send result/error to channel
capture current loop value
call multiple functions
add logging/metrics around work
```

Example:

```go
for _, job := range jobs {
    job := job // important before Go 1.22
    wg.Add(1)

    go func() {
        defer wg.Done()

        result, err := process(job)
        if err != nil {
            errCh <- err
            return
        }

        resultCh <- result
    }()
}
```

Use a named function when the function already has everything it needs and you do not need wrapper logic.

```go
func worker(ctx context.Context, jobs <-chan Job, results chan<- Result) {
    for {
        select {
        case <-ctx.Done():
            return
        case job, ok := <-jobs:
            if !ok {
                return
            }
            results <- process(job)
        }
    }
}

go worker(ctx, jobs, results)
```

Avoid anonymous functions when they become large.

```go
// Avoid this shape.
go func() {
    // 80 lines of business logic
}()

// Prefer this shape.
go processUserJob(ctx, user, resultCh)
```

Important loop-variable pattern:

```go
for _, user := range users {
    user := user
    go func() {
        sendEmail(user)
    }()
}
```

Or pass the value as an argument:

```go
for _, user := range users {
    go func(u User) {
        sendEmail(u)
    }(user)
}
```

Interview wording:

> I use an anonymous goroutine when I need small wrapper logic like `defer wg.Done`, error handling, result sending, context select, or capturing loop variables. If the goroutine body is meaningful business logic or reused, I move it into a named function and call it with `go functionName(...)`.

## Multiple Correct Answers

Many Go concurrency interview questions have more than one valid solution.

Question:

> Process 100 URLs concurrently.

Option 1:

```text
small bounded list -> one goroutine per URL + WaitGroup
```

Option 2:

```text
large list or API limit -> worker pool
```

Option 3:

```text
need first success only -> result channel + context cancellation
```

Option 4:

```text
need clean error handling -> errgroup.WithContext
```

Senior-level answer:

> If the number of tasks is small and bounded, I can start one goroutine per task and wait with a WaitGroup. If the task count is large or the dependency has limits, I use a worker pool. If I need results, I use channels or errgroup. If the work is request-scoped or should stop on timeout, I pass context.

## Practical Decision Table

| Situation | Good Tool |
|---|---|
| Run independent background work | goroutine |
| Wait for N goroutines | WaitGroup |
| Pass jobs/results between goroutines | channel |
| Stop work on timeout/cancel/shutdown | context |
| Protect shared map/counter/cache | mutex/RWMutex |
| Process many tasks with a fixed limit | worker pool |
| Wait for work, timeout, or cancel | select |
| Collect errors from goroutines | errgroup |
| Stream many values through stages | channels/pipeline |

## One-Line Interview Summary

> In Go, I use goroutines to run work concurrently, channels to communicate between goroutines, WaitGroup to wait for known goroutines, context for cancellation and timeouts, mutexes for shared state, worker pools to limit concurrency, and select to coordinate multiple channel events.

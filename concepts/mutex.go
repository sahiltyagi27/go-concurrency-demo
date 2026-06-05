// CONCEPT 8: Mutex and RWMutex
//
// Channels are great for ownership transfer and pipelines.
// Mutexes are great when several goroutines must safely access the same memory.
//
// sync.Mutex:
//   One goroutine at a time can enter the critical section.
//
// sync.RWMutex:
//   Many readers can hold RLock at once, but writers need exclusive Lock.
//   Use it when reads are frequent and writes are less frequent.

package concepts

import (
	"fmt"
	"sync"
)

// SafeCounter protects shared state with a Mutex.
type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// ScoreCache demonstrates RWMutex for read-heavy shared state.
type ScoreCache struct {
	mu     sync.RWMutex
	scores map[string]int
}

func NewScoreCache() *ScoreCache {
	return &ScoreCache{scores: make(map[string]int)}
}

func (c *ScoreCache) Set(source string, score int) {
	c.mu.Lock() // exclusive writer lock
	defer c.mu.Unlock()
	c.scores[source] = score
}

func (c *ScoreCache) Get(source string) (int, bool) {
	c.mu.RLock() // shared reader lock
	defer c.mu.RUnlock()
	score, ok := c.scores[source]
	return score, ok
}

func RunSharedStateDemo() {
	var wg sync.WaitGroup
	counter := &SafeCounter{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				counter.Inc()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("SafeCounter final value: got=%d want=5000\n", counter.Value())

	cache := NewScoreCache()
	cache.Set("TechCrunch", 91)
	cache.Set("HackerNews", 88)

	var readers sync.WaitGroup
	for i := 0; i < 3; i++ {
		readers.Add(1)
		go func(id int) {
			defer readers.Done()
			if score, ok := cache.Get("TechCrunch"); ok {
				fmt.Printf("reader %d saw TechCrunch score=%d\n", id, score)
			}
		}(i)
	}
	readers.Wait()
}

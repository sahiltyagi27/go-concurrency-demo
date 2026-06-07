// CONCEPT 10: sync.WaitGroup.Go
//
// Go 1.25 added WaitGroup.Go as a convenience method.
// It starts a goroutine, increments the WaitGroup counter before the goroutine
// runs, and decrements the counter when the goroutine returns.
//
// It replaces this common pattern:
//
//   wg.Add(1)
//   go func() {
//       defer wg.Done()
//       work()
//   }()
//
// with:
//
//   wg.Go(func() {
//       work()
//   })
//
// Interview note:
//   WaitGroup.Go reduces the chance of common mistakes like calling Add inside
//   the goroutine or forgetting Done. The classic Add/Done pattern is still
//   important because older Go versions and many codebases use it.

package concepts

import (
	"fmt"
	"sync"
)

func RunWaitGroupGoDemo() {
	var wg sync.WaitGroup
	results := make(chan string, 2)

	wg.Go(func() {
		results <- "profile loaded"
	})

	wg.Go(func() {
		results <- "orders loaded"
	})

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Println(result)
	}
}

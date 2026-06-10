package concepts

import "fmt"

// RunPingPongDemo alternates two goroutines using unbuffered channels.
func RunPingPongDemo() {
	pingCh := make(chan struct{})
	pongCh := make(chan struct{})
	done := make(chan struct{})

	const n = 5

	go func() {
		for i := 0; i < n; i++ {
			<-pingCh
			fmt.Println("ping")
			pongCh <- struct{}{}
		}
	}()

	go func() {
		for i := 0; i < n; i++ {
			<-pongCh
			fmt.Println("pong")
			if i < n-1 {
				pingCh <- struct{}{}
			}
		}
		done <- struct{}{}
	}()

	pingCh <- struct{}{}
	<-done
}

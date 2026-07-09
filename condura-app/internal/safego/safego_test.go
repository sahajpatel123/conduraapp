package safego

import (
	"sync"
	"testing"
	"time"
)

func TestGo_NilNoPanic(t *testing.T) {
	Go(nil)
}

func TestGo_RecoversPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Go(func() {
		defer wg.Done()
		panic("intentional safego test panic")
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		// crash.Recover ate the panic; process still alive.
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for safego.Go panic recovery")
	}
}

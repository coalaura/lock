package lock

import (
	"sync"
	"testing"
)

func TestSpinLock_Basic(t *testing.T) {
	l := NewLock()

	// Test TryLock
	if !l.TryLock() {
		t.Fatal("expected TryLock to succeed on fresh lock")
	}

	if l.TryLock() {
		t.Fatal("expected TryLock to fail when already locked")
	}

	l.Unlock()

	if !l.TryLock() {
		t.Fatal("expected TryLock to succeed after Unlock")
	}

	l.Unlock()
}

func TestSpinLock_Concurrency(t *testing.T) {
	l := NewLock()

	var (
		counter int
		wg      sync.WaitGroup
	)

	for range 1000 {
		wg.Go(func() {
			l.Lock()
			counter++
			l.Unlock()
		})
	}

	wg.Wait()

	if counter != 1000 {
		t.Fatalf("expected counter to be 1000, got %d", counter)
	}
}

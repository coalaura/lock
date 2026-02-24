package lock

import (
	"sync"
	"testing"
)

func TestLockMap_BasicAndCleanup(t *testing.T) {
	m := NewLockMap[string]()

	m.Lock("A")

	if len(m.locks) != 1 {
		t.Fatalf("expected 1 lock in map, got %d", len(m.locks))
	}

	m.Unlock("A")

	if len(m.locks) != 0 {
		t.Fatalf("expected map to be empty after unlock, got %d", len(m.locks))
	}
}

func TestLockMap_Isolation(t *testing.T) {
	m := NewLockMap[string]()

	m.Lock("A")

	if !m.TryLock("B") {
		t.Fatal("expected to be able to lock B while A is locked")
	}

	if m.TryLock("A") {
		t.Fatal("expected TryLock on A to fail")
	}

	m.Unlock("A")
	m.Unlock("B")
}

func TestLockMap_Concurrency(t *testing.T) {
	m := NewLockMap[int]()

	var wg sync.WaitGroup

	counters := make([]int, 3)

	for range 100 {
		wg.Go(func() {
			for range 50 {
				m.Lock(1)
				counters[1]++
				m.Unlock(1)

				m.Lock(2)
				counters[2]++
				m.Unlock(2)
			}
		})
	}

	wg.Wait()

	if counters[1] != 5000 || counters[2] != 5000 {
		t.Fatalf("expected counters to be 5000, got %v", counters)
	}

	if len(m.locks) != 0 {
		t.Fatalf("memory leak: expected locks map to be empty, got %d", len(m.locks))
	}
}

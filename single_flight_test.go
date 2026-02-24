package lock

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSingleFlightMap_SingleExecution(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	var (
		calls int32
		wg    sync.WaitGroup
	)

	for range 100 {
		wg.Go(func() {
			val, err := m.GetOrFill("key1", func() (string, error) {
				atomic.AddInt32(&calls, 1)

				time.Sleep(50 * time.Millisecond) // work

				return "generated", nil
			})

			if err != nil || val != "generated" {
				t.Errorf("unexpected result: %v, %v", val, err)
			}
		})
	}

	wg.Wait()

	if calls != 1 {
		t.Fatalf("expected generator to be called exactly 1 time, got %d", calls)
	}

	if len(m.calls) != 0 {
		t.Fatal("memory leak: calls map should be empty")
	}
}

func TestSingleFlightMap_TTLAndMemoryCleanup(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	m.SetWithTTL("key1", "val", 50*time.Millisecond)

	if _, ok := m.Get("key1"); !ok {
		t.Fatal("expected key to exist")
	}

	if len(m.timers) != 1 {
		t.Fatal("expected 1 active timer")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := m.Get("key1"); ok {
		t.Fatal("expected key to be deleted by TTL")
	}

	if len(m.items) != 0 || len(m.timers) != 0 {
		t.Fatalf("memory leak: expected items and timers to be empty. items:%d timers:%d", len(m.items), len(m.timers))
	}
}

func TestSingleFlightMap_StaleOverwriteProtection(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	var wg sync.WaitGroup

	wg.Go(func() {
		m.GetOrFill("key1", func() (string, error) {
			time.Sleep(100 * time.Millisecond)

			return "STALE_DATA", nil
		})
	})

	time.Sleep(20 * time.Millisecond)

	m.Set("key1", "FRESH_DATA")

	wg.Wait()

	val, _ := m.Get("key1")
	if val != "FRESH_DATA" {
		t.Fatalf("expected FRESH_DATA, got %s", val)
	}
}

func TestSingleFlightMap_GetAndWait(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	if _, ok := m.GetAndWait("key1"); ok {
		t.Fatal("expected GetAndWait to return false for missing item")
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		m.GetOrFill("key1", func() (string, error) {
			time.Sleep(100 * time.Millisecond)

			return "slow_val", nil
		})
	})

	time.Sleep(20 * time.Millisecond)

	start := time.Now()
	val, ok := m.GetAndWait("key1")
	duration := time.Since(start)

	if !ok || val != "slow_val" {
		t.Fatalf("expected to get slow_val, got %v (ok: %v)", val, ok)
	}

	if duration < 50*time.Millisecond {
		t.Fatalf("expected GetAndWait to block, but it returned instantly")
	}

	wg.Wait()
}

func TestSingleFlightMap_GeneratorError(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	_, err := m.GetOrFill("key1", func() (string, error) {
		return "", errors.New("generation failed")
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if _, ok := m.Get("key1"); ok {
		t.Fatal("expected key to NOT exist after generator error")
	}
}

func TestSingleFlightMap_DeleteCancelsTimer(t *testing.T) {
	m := NewSingleFlightMap[string, string]()

	m.SetWithTTL("key1", "val", 1*time.Hour)

	if len(m.timers) != 1 {
		t.Fatal("expected timer to exist")
	}

	m.Delete("key1")

	if len(m.timers) != 0 {
		t.Fatal("expected timer to be deleted after manual Delete")
	}
}

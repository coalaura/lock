package lock

import (
	"sync"
)

type refCounter struct {
	mu    sync.Mutex
	count int
}

// LockMap provides independent, concurrent-safe locks for specific keys.
type LockMap[T comparable] struct {
	mx    sync.Mutex
	locks map[T]*refCounter
}

func NewLockMap[T comparable]() *LockMap[T] {
	return &LockMap[T]{
		locks: make(map[T]*refCounter),
	}
}

// Lock acquires a blocking lock for the specific key.
func (m *LockMap[T]) Lock(key T) {
	m.mx.Lock()

	ref, ok := m.locks[key]
	if !ok {
		ref = &refCounter{}

		m.locks[key] = ref
	}

	ref.count++
	m.mx.Unlock()

	ref.mu.Lock()
}

// TryLock attempts to acquire the lock for the specific key without blocking.
// Returns true if acquired, false if another goroutine currently holds it.
func (m *LockMap[T]) TryLock(key T) bool {
	m.mx.Lock()

	ref, ok := m.locks[key]
	if !ok {
		ref = &refCounter{}
		m.locks[key] = ref
	}

	acquired := ref.mu.TryLock()
	if acquired {
		ref.count++
	} else if !ok {
		delete(m.locks, key)
	}

	m.mx.Unlock()

	return acquired
}

// Unlock releases the lock for the specific key and cleans up memory
// if no other goroutines are waiting for it.
func (m *LockMap[T]) Unlock(key T) {
	m.mx.Lock()

	ref, ok := m.locks[key]
	if !ok {
		m.mx.Unlock()

		panic("unlock of unlocked key")
	}

	ref.count--

	if ref.count == 0 {
		delete(m.locks, key)
	}

	m.mx.Unlock()

	ref.mu.Unlock()
}

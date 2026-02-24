package lock

import (
	"runtime"
	"sync/atomic"
)

// SpinLock is a lightweight, atomic lock. It is highly efficient for
// extremely short critical sections where putting a goroutine to sleep
// (like sync.Mutex does) would be too slow.
type SpinLock struct {
	locked atomic.Bool
}

func NewLock() *SpinLock {
	return &SpinLock{}
}

// TryLock attempts to acquire the lock and returns true if successful.
// It returns immediately and does not block.
func (l *SpinLock) TryLock() bool {
	return l.locked.CompareAndSwap(false, true)
}

// Lock blocks until the lock is acquired.
func (l *SpinLock) Lock() {
	for !l.TryLock() {
		runtime.Gosched()
	}
}

// Unlock releases the lock.
func (l *SpinLock) Unlock() {
	l.locked.Store(false)
}

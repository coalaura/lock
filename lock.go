package lock

import "sync/atomic"

type Lock struct {
	lock uintptr
}

func NewLock() *Lock {
	return &Lock{}
}

func (l *Lock) Lock() bool {
	return atomic.CompareAndSwapUintptr(&l.lock, 0, 1)
}

func (l *Lock) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}

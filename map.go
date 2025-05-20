package lock

import (
	"sync"
)

type LockMap[T comparable] struct {
	mx    sync.Mutex
	locks map[T]*Lock
}

func NewLockMap[T comparable]() *LockMap[T] {
	return &LockMap[T]{
		locks: make(map[T]*Lock, 0),
	}
}

func (m *LockMap[T]) Lock(key T) bool {
	m.mx.Lock()

	lock, ok := m.locks[key]
	if !ok {
		lock = NewLock()

		m.locks[key] = lock
	}

	m.mx.Unlock()

	return lock.Lock()
}

func (m *LockMap[T]) Unlock(key T) {
	m.mx.Lock()
	defer m.mx.Unlock()

	lock, ok := m.locks[key]
	if !ok {
		return
	}

	lock.Unlock()

	delete(m.locks, key)
}

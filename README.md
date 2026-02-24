# lock

A collection of optimized, specialized synchronization primitives and concurrent data structures for Go.

```bash
go get github.com/coalaura/lock
```

## Primitives

### 1. SpinLock
A lightweight, atomic lock. It avoids putting goroutines to sleep, making it highly efficient for extremely short critical sections where `sync.Mutex` overhead is too high.

```go
l := lock.NewLock()

l.Lock()
// ... short critical section ...
l.Unlock()

// Non-blocking alternative:
if l.TryLock() {
    defer l.Unlock()
}
```

### 2. LockMap
Provides independent, concurrent-safe locks for specific keys. It automatically cleans up memory (reference counting) when no goroutines are waiting for a key.

```go
m := lock.NewLockMap[string]()

// Goroutines will only block if they request the exact same key
m.Lock("user_123")
// ... modify user_123 ...
m.Unlock("user_123")
```

### 3. SingleFlightMap
A highly optimized concurrent cache with built-in SingleFlight coordination and optional TTLs. It guarantees that a generator function is only executed once for a missing key, preventing cache stampedes.

```go
cache := lock.NewSingleFlightMap[string, *User]()

// If 100 goroutines call this simultaneously, the database logic 
// only runs once. The other 99 wait and share the result.
user, err := cache.GetOrFillWithTTL("user_123", 5*time.Minute, func() (*User, error) {
    return fetchUserFromDB("user_123")
})
```

*Additional `SingleFlightMap` features:*
* `Get(key)`: Fast, non-blocking read.
* `GetAndWait(key)`: Wait for in-flight generations without triggering one.
* `Set(key, val)` / `SetWithTTL(key, val, ttl)`: Forceful overwrites with stale-data protection.
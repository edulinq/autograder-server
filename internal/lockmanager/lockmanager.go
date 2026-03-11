package lockmanager

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

// All fields in lockData are only modified while holding lockManagerMutex,
// except for mutex (which is the lock itself).
// timestamp and lockCount use atomic types to allow safe reads outside the mutex
// (e.g., the initial staleness check in RemoveStaleLocksOnce).
type lockData struct {
	timestamp atomic.Int64
	mutex     sync.RWMutex
	lockCount atomic.Int64
}

const TICKER_DURATION_HOUR = 1.0

var (
	lockManagerMutex sync.Mutex
	lockMap          sync.Map
	ticker           *time.Ticker
)

func init() {
	ticker = time.NewTicker(TICKER_DURATION_HOUR * time.Hour)
	go removeStaleLocks()
}

// Acquire a write lock on the given key.
// The returned boolean indicates if the lock was likely acquired without waiting
// (e.g. true if the thread did not have to wait and false if there was another thread holding the lock).
// This boolean return will not always be correct and should only be used as an advisement, not a hard fact.
func Lock(key string) bool {
	return lock(key, false)
}

func Unlock(key string) error {
	return unlock(key, false)
}

func ReadLock(key string) {
	lock(key, true)
}

func ReadUnlock(key string) error {
	return unlock(key, true)
}

func lock(key string, read bool) bool {
	lockManagerMutex.Lock()

	val, _ := lockMap.LoadOrStore(key, &lockData{})
	lock := val.(*lockData)

	// Note if any other threads currently have this lock.
	lockNotInUse := (lock.lockCount.Load() == 0)

	// Update the lock state before releasing lockManagerMutex.
	// This ensures RemoveStaleLocksOnce() always sees consistent, up-to-date state
	// when it inspects entries without holding the mutex.
	lock.lockCount.Add(1)
	lock.timestamp.Store(timestamp.Now().ToMSecs())

	// Unlock the lockManagerMutex before acquiring the lock to avoid a deadlock.
	lockManagerMutex.Unlock()

	if read {
		lock.mutex.RLock()
	} else {
		lock.mutex.Lock()
	}

	log.Trace("Lock", log.NewAttr("read", read), log.NewAttr("key", key))

	return lockNotInUse
}

func unlock(key string, read bool) error {
	lockManagerMutex.Lock()
	defer lockManagerMutex.Unlock()

	val, exists := lockMap.Load(key)
	if !exists {
		log.Error("Key does not exist.", log.NewAttr("key", key))
		return fmt.Errorf("Lock key not found: '%s'.", key)
	}

	lock := val.(*lockData)
	if !read && lock.lockCount.Load() <= 0 {
		log.Error("Tried to unlock a lock that is already unlocked: %s\n", key)
		return fmt.Errorf("Tried to unlock a lock that is already unlocked with key '%s'", key)
	}

	lock.lockCount.Add(-1)
	lock.timestamp.Store(timestamp.Now().ToMSecs())

	log.Trace("Unlock", log.NewAttr("read", read), log.NewAttr("key", key))

	if read {
		lock.mutex.RUnlock()
	} else {
		lock.mutex.Unlock()
	}

	return nil
}

func removeStaleLocks() {
	for range ticker.C {
		RemoveStaleLocksOnce()
	}
}

func RemoveStaleLocksOnce() {
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second
	staleDurationMSecs := staleDuration.Milliseconds()

	lockMap.Range(func(key, val any) bool {
		lock := val.(*lockData)

		// First, check if the lock isn't stale or is locked.
		if (timestamp.Now().ToMSecs()-lock.timestamp.Load() < staleDurationMSecs) || (lock.lockCount.Load() > 0) {
			return true
		}

		// Lock the lock manager in case another thread is trying to lock/unlock.
		lockManagerMutex.Lock()
		defer lockManagerMutex.Unlock()

		// Second, try to acquire the lock.
		if lock.mutex.TryLock() {
			defer lock.mutex.Unlock()

			// Finally, if the lock is stale, delete it.
			if timestamp.Now().ToMSecs()-lock.timestamp.Load() > staleDurationMSecs {
				lockMap.Delete(key)
			}
		}

		return true
	})
}

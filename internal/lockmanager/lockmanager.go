package lockmanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
)

type lockData struct {
	timestamp time.Time
	mutex     sync.RWMutex
	lockCount int
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

func Lock(key string) {
	lock(key, false)
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

func lock(key string, read bool) {
	lockManagerMutex.Lock()

	val, _ := lockMap.LoadOrStore(key, &lockData{})
	lock := val.(*lockData)

	// Unlock the lockManagerMutex before acquiring the lock to avoid a deadlock.
	lockManagerMutex.Unlock()

	if read {
		lock.mutex.RLock()
	} else {
		lock.mutex.Lock()
	}

	log.Trace("Lock", log.NewAttr("read", read), log.NewAttr("key", key))

	lock.lockCount++
	lock.timestamp = time.Now()
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
	if !read && lock.lockCount == 0 {
		log.Error("Tried to unlock a lock that is already unlocked: %s\n", key)
		return fmt.Errorf("Tried to unlock a lock that is already unlocked with key '%s'", key)
	}

	lock.lockCount--
	lock.timestamp = time.Now()

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
	staleDuration := time.Duration(time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second)

	lockMap.Range(func(key, val any) bool {
		lock := val.(*lockData)

		// First check: If the lock isn't stale or is locked, return early.
		if time.Since(lock.timestamp) < staleDuration || lock.lockCount > 0 {
			return true
		}

		// Lock the lock manager in case another thread is trying to lock/unlock.
		lockManagerMutex.Lock()
		defer lockManagerMutex.Unlock()

		// Second check: If the lock is stale and and is able to be locked, delete it.
		if time.Since(lock.timestamp) > staleDuration && lock.mutex.TryLock() {
			lockMap.Delete(key)
		}

		return true
	})
}

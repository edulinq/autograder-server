package common

import (
    "fmt"
    "sync"
    "time"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/log"
)

type lockData struct {
    timestamp time.Time;
    mutex sync.RWMutex;
    isLocked bool;
}

const TICKER_DURATION_HOUR = 1.0;

var (
    lockManagerMutex sync.Mutex;
    lockMap sync.Map;
    ticker *time.Ticker;
)

func init() {
    ticker = time.NewTicker(TICKER_DURATION_HOUR * time.Hour);
    go removeStaleLocks();
}

func Lock(key string) {
    lockManagerMutex.Lock();
    defer lockManagerMutex.Unlock();

    val, _ := lockMap.LoadOrStore(key, &lockData{});
    lock := val.(*lockData)

    lock.mutex.Lock();	
    lock.timestamp = time.Now();
    lock.isLocked = true;
}

func Unlock(key string) error {
    lockManagerMutex.Lock();
    defer lockManagerMutex.Unlock();

    val, exists := lockMap.Load(key);
    if !exists {
        log.Error("Key does not exist.", log.NewAttr("key", key));
        return fmt.Errorf("Lock key not found: '%s'.", key);
    }
    
    lock := val.(*lockData);
    if !lock.isLocked {
        log.Error("Tried to unlock a lock that is unlocked with key.", log.NewAttr("key", key));
        return fmt.Errorf("Tried to unlock a lock that is already unlocked with key '%s'.", key);
    }

    lock.isLocked = false;
    lock.timestamp = time.Now();
    lock.mutex.Unlock();

    return nil;
}

func removeStaleLocks() {
    for range ticker.C {
        RemoveStaleLocksOnce();
    }
}

func RemoveStaleLocksOnce() {
    staleDuration := time.Duration(time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second);

    lockMap.Range(func(key, val any) bool {
        lock := val.(*lockData);

        if time.Since(lock.timestamp) < staleDuration || lock.isLocked {
            return true;
        }

        lockManagerMutex.Lock();
        defer lockManagerMutex.Unlock();
        if time.Since(lock.timestamp) > staleDuration && lock.mutex.TryLock() {
            lockMap.Delete(key);
        }

        return true;
    })
}

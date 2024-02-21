package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/eriq-augustine/autograder/config"
	"github.com/eriq-augustine/autograder/log"
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
)

func init() {
    go removeStaleLocks();
}

func Lock(key string) {
    lockManagerMutex.Lock();
    defer lockManagerMutex.Unlock();

    val, _ := lockMap.LoadOrStore(key, &lockData{});
    val.(*lockData).mutex.Lock();	

    val.(*lockData).timestamp = time.Now();
    val.(*lockData).isLocked = true;
}

func Unlock(key string) error {
    lockManagerMutex.Lock();
    defer lockManagerMutex.Unlock();

    val, exists := lockMap.Load(key);
    if !exists {
        log.Error("Key does not exist.", log.NewAttr("key", key));
        return fmt.Errorf("Key not found: '%s'.", key);
    }
    
    lock := val.(*lockData);
    if !lock.isLocked {
        log.Error("Tried to unlock a lock that is unlocked with key.", log.NewAttr("key", key));
        return fmt.Errorf("Key '%s' tried to unlock a lock that is unlocked.", key);
    }

    lock.isLocked = false;
    lock.timestamp = time.Now();	
    lock.mutex.Unlock();
    return nil;
}


func removeStaleLocks() {
    ticker := time.NewTicker(TICKER_DURATION_HOUR * time.Hour);

    for range ticker.C {
        RemoveStaleLocksOnce();
    }
}

func RemoveStaleLocksOnce() {
    staleDuration := time.Duration(time.Duration(config.STALELOCK_DURATION.Get()) * time.Second);
    
    lockMap.Range(func(key, val any) bool {
        lock := val.(*lockData);

        if (time.Since(lock.timestamp) > staleDuration) && (!lock.isLocked) {
            lockManagerMutex.Lock();
            defer lockManagerMutex.Unlock();
            if (time.Since(lock.timestamp) > staleDuration) && (lock.mutex.TryLock()) {
                lockMap.Delete(key);
            }
        }
        return true;
    })
}
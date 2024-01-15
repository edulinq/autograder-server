package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/eriq-augustine/autograder/config"
	"github.com/rs/zerolog/log"
)

type lockData struct {
	timestamp time.Time;
	mutex sync.Mutex;
	isLocked bool;
}

var (
	lockManagerMutex sync.Mutex;
	lockMap sync.Map;
	staleDuration time.Duration;
)

func init() {
	go removeStaleLocks();
}

func Lock(key string) {
	val, _ := lockMap.LoadOrStore(key, &lockData{});
	val.(*lockData).mutex.Lock();	

	lockManagerMutex.Lock();
	defer lockManagerMutex.Unlock()

	val.(*lockData).timestamp = time.Now();
	val.(*lockData).isLocked = true;

}

func Unlock(key string) error {
	
	val, exists := lockMap.Load(key);
	lock := val.(*lockData);
	if !exists {
		log.Error().Str("key", key).Msg("Key does not exist");
		return fmt.Errorf("Key not found: %v", key);
	}

	lockManagerMutex.Lock();
	defer lockManagerMutex.Unlock();

	if lock.isLocked {
		lock.isLocked = false;
		defer lock.mutex.Unlock();
		lock.timestamp = time.Now();	
	} else {
		log.Error().Str("key", key).Msg("Tried to unlock a lock that is unlocked");
		return fmt.Errorf("Key: %v Tried to unlock a lock that is unlocked", key);
	}

	return nil;
}

func removeStaleLocks() {
	ticker := time.NewTicker(time.Duration(config.STALELOCK_DURATION) * time.Second);
	defer ticker.Stop();

	for range ticker.C {
		staleDuration := time.Duration(config.STALELOCK_DURATION) * time.Second;

        lockMap.Range(func(key, val any) bool {
            lock := val.(*lockData);
            if (time.Since(lock.timestamp) > staleDuration) && (lock.mutex.TryLock()) {
				lock.mutex.Unlock();
				lockManagerMutex.Lock();
				if (time.Since(lock.timestamp) > staleDuration) && (lock.mutex.TryLock()) {
					lockMap.Delete(key);
				}
				lockManagerMutex.Unlock();
            }
            return true;
        })
    }
}
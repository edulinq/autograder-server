package util

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Lock struct {
	key string;
	timestamp time.Time;
}

type LockManager struct {
	locks map[string]*Lock;
	mutex sync.Mutex;
	staleDuration time.Duration;
}

func NewLockManager(staleDuration time.Duration) *LockManager {
	lm := &LockManager{
		locks: make(map[string]*Lock),
		staleDuration: staleDuration,
	}
	go lm.removeStaleLocks();
	return lm;
}

func (lm *LockManager) Lock(key string) {
	lm.mutex.Lock();
	lm.locks[key] = &Lock{key: key, timestamp: time.Now()};
}

func (lm *LockManager) Unlock(key string) error {
	// Check if key exists in the lock map
	_, ok := lm.locks[key];
	if ok {
		defer lm.mutex.Unlock();
		delete(lm.locks, key);
		return nil;
	} else {
		log.Fatal("Key not found.")
		return fmt.Errorf("Error. Key not found.");
	}
}

func (lm *LockManager) removeStaleLocks() {
	time.Sleep(lm.staleDuration);
	for key, lock := range lm.locks {
		if time.Since(lock.timestamp) > lm.staleDuration {
			delete(lm.locks, key);
			lm.mutex.Unlock();
		}
	}
}
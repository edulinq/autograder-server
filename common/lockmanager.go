package common

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/eriq-augustine/autograder/config"
)

type lockInfo struct {
	key string;
	timestamp time.Time;
}

type lockManager struct {
	lmMutex sync.Mutex;
	lockMutex sync.Map;
	lockInstances map[string]*lockInfo;
	staleDuration time.Duration;
}

func NewLockManager() *lockManager {
	lm := &lockManager{
		lockInstances: make(map[string]*lockInfo),
		staleDuration: time.Duration(config.STALELOCK_DURATION) * time.Second,
	}
	go lm.removeStaleLocks();
	return lm;
}

func (lm *lockManager) Lock(key string) {
	val, _ := lm.lockMutex.LoadOrStore(key, &sync.Mutex{});
	val.(*sync.Mutex).Lock(); // Lock the mutex for the associated key.
	
	// Only one go routine can write to the lockInstance map at a time.
	lm.lmMutex.Lock(); 
	lm.lockInstances[key] = &lockInfo{key: key, timestamp: time.Now()};
	lm.lmMutex.Unlock();
}

func (lm *lockManager) Unlock(key string) error {
	val, ok := lm.lockMutex.Load(key);
	if !ok {
		log.Printf("Error. Key not found: %v", key);
		return fmt.Errorf("Error. Key not found: %v", key);
	}

	defer val.(*sync.Mutex).Unlock(); // Unlock the mutex for the associated key.

	// Only one go routine can delete a key from the lockInstance map at a time.
	lm.lmMutex.Lock();
	delete(lm.lockInstances, key);
	lm.lmMutex.Unlock();

	return nil;
}

func (lm *lockManager) removeStaleLocks() {
	ticker := time.NewTicker(lm.staleDuration);
	
	for range ticker.C {
		lm.lmMutex.Lock();
		for key, timestamp := range lm.lockInstances {
			if time.Since(timestamp.timestamp) > lm.staleDuration {
				val, _ := lm.lockMutex.Load(key);
				val.(*sync.Mutex).Unlock();
				delete(lm.lockInstances, key);
			}
		}
		lm.lmMutex.Unlock();
	}
}
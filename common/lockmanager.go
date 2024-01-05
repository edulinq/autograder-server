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
	mutex sync.Mutex;
	lockMutex map[string]*sync.Mutex;
	lockInstances map[string]*lockInfo;
	staleDuration time.Duration;
}

func NewLockManager() *lockManager {
	lm := &lockManager{
		lockMutex: make(map[string]*sync.Mutex),
		lockInstances: make(map[string]*lockInfo),
		staleDuration: time.Duration(config.STALELOCK_DURATION) * time.Second,
	}
	go lm.removeStaleLocks();
	return lm;
}

func (lm *lockManager) Lock(key string) {
	// Only one go routine can read/write to the lockMutex map at a time.
	lm.mutex.Lock();
	
	if _, exists := lm.lockMutex[key]; !exists {
		lm.lockMutex[key] = &sync.Mutex{}; // Each unique key gets a mutex.
	}
	
	mutex := lm.lockMutex[key];
	lm.mutex.Unlock();
	mutex.Lock(); // Lock the mutex for the associated key.
	

	// Only one go routine can write to the lockInstance map at a time.
	lm.mutex.Lock(); 
	lm.lockInstances[key] = &lockInfo{key: key, timestamp: time.Now()};
	lm.mutex.Unlock();
}

func (lm *lockManager) Unlock(key string) error {
	// Only one go routine can read/delete to the lockMutex map at a time.
	lm.mutex.Lock();
	mutex, ok := lm.lockMutex[key];
	if ok {
		delete(lm.lockInstances, key);
	}
	lm.mutex.Unlock();

	if ok {
		mutex.Unlock(); // Unlock the mutex witht he associated key if the key is found in the map.
		return nil;
	} else {
		log.Printf("Error. Key not found: %v", key);
		return fmt.Errorf("Error. Key not found: %v", key);
	}
}

func (lm *lockManager) removeStaleLocks() {
	ticker := time.NewTicker(lm.staleDuration);
	for range ticker.C{
		lm.mutex.Lock();
		fmt.Println("Tick")
		for key, timestamp := range lm.lockInstances {
			// fmt.Println("time since gotten the lock: ", time.Since(timestamp.timestamp))
			// fmt.Println("staleDuration: ", lm.staleDuration)
			if time.Since(timestamp.timestamp) > lm.staleDuration {
				mutex, _ := lm.lockMutex[key];
				mutex.Unlock();
				delete(lm.lockMutex, key);
				delete(lm.lockInstances, key);
			}
		}
		lm.mutex.Unlock();
	}
}
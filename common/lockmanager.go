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
	val, _ := lm.lockMutex.LoadOrStore(key, &sync.Mutex{})
	mutex := val.(*sync.Mutex);
	mutex.Lock(); // Lock the mutex for the associated key.
	
	// Only one go routine can write to the lockInstance map at a time.
	lm.lmMutex.Lock(); 
	lm.lockInstances[key] = &lockInfo{key: key, timestamp: time.Now()};
	lm.lmMutex.Unlock();
	// fmt.Println("here")
}

func (lm *lockManager) Unlock(key string) error {
	// fmt.Println(lm.lockMutex.Load(key))
	val, ok := lm.lockMutex.Load(key);
	// fmt.Println("Ok: ", ok)
	if !ok {
		log.Printf("Error. Key not found: %v", key);
		return fmt.Errorf("Error. Key not found: %v", key);
	}
	mutex := val.(*sync.Mutex);
	// Only one go routine can read/delete to the lockMutex map at a time.
	lm.lmMutex.Lock();
	delete(lm.lockInstances, key);
	lm.lmMutex.Unlock();

	mutex.Unlock();
	return nil;
}

func (lm *lockManager) removeStaleLocks() {
	ticker := time.NewTicker(lm.staleDuration);
	for range ticker.C{
		lm.lmMutex.Lock();
		fmt.Println("Tick")
		for key, timestamp := range lm.lockInstances {
			if time.Since(timestamp.timestamp) > lm.staleDuration {
				val, _ := lm.lockMutex.Load(key);
				mutex := val.(*sync.Mutex);
				mutex.Unlock();
				delete(lm.lockInstances, key);
			}
		}
		lm.lmMutex.Unlock();
	}
}
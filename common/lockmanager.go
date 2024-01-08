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
}

var (
	lockMap sync.Map;
	staleDuration time.Duration;
)

func init() {
	staleDuration = time.Duration(config.STALELOCK_DURATION) * time.Second;
	go removeStaleLocks();
}

func Lock(key string) {
	val, _ := lockMap.LoadOrStore(key, &lockData{});
	val.(*lockData).mutex.Lock();	
	val.(*lockData).timestamp = time.Now();
}

func Unlock(key string) error {
	_, exists := lockMap.LoadAndDelete(key);
	if !exists {
		log.Error().Str("key", key).Msg("Key does not exist");
		return fmt.Errorf("Error. Key not found: %v", key);
	}

	return nil;
}

func removeStaleLocks() {
	ticker := time.NewTicker(staleDuration);

	for range ticker.C {
        lockMap.Range(func(key, val interface{}) bool {
            lock := val.(*lockData);
            if time.Since(lock.timestamp) > staleDuration {
				lock.mutex.Unlock();                
            }
            return true;
        })
    }
}
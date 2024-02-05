package common

import (
	"sync"
	"testing"
	"time"
)

var (
	removeKey = "remove"
	key = "testkey1"
	key2 = "testkey2"
	doesNotExistKey = "dne"
)

func TestLock(test *testing.T) {
    var wg sync.WaitGroup;

	// Testing Lock
	Lock(key);

	// Testing Unlock
	err := Unlock(key);
	if (err != nil) {
		test.Errorf("Failed to unlock a key");
	}

	// Testing remove stale locks
    Lock(removeKey);
    Unlock(removeKey);
	wg.Add(1)
	go func() {
		defer wg.Done();
		time.Sleep(2 * time.Second);
		if !RemoveStaleLocksOnce() {
			test.Errorf("Failed to remove stale locks");
		} 
	}()

	// Testing concurrent Locking/Unlocking
    Lock(key);
    Lock(key2);
	wg.Add(2);
    go func() {
        defer wg.Done();
        Lock(key);
        defer Unlock(key);
    }()
    go func() {
        defer wg.Done();
        Lock(key2);
        defer Unlock(key2);
    }();
    Unlock(key);
    Unlock(key2);

    wg.Wait();

	// Testing Unlocking an unlocked lock
	err = Unlock(key);
	if err == nil {
		test.Errorf("Lockmanager unlocked an unlocked key");
	}
	
	// Testing Unlocking a key that doesn't exist
	err = Unlock(doesNotExistKey);
	if err == nil {
		test.Errorf("Lockmanager unlocked a key that doesn't exist");
	}

}
package common

import (
	"sync"
	"testing"
	"time"

	"github.com/eriq-augustine/autograder/config"
)

var (
    key = "testkey1"
    key2 = "testkey2"
    doesNotExistKey = "dne"
    staleDuration = time.Duration(config.STALELOCK_DURATION.Get()) * time.Second;
	wg sync.WaitGroup;

)

func TestLock(test *testing.T) {
	testLockingKey()

	testUnlockingKey(test)

	testConcurrentLockingUnlocking(test)

    testUnlockingAnUnlockedLock(test)
    
	testUnlockingMissingKey(test)

	testLockConcurrencyWithStaleCheck(test)
}

func testLockingKey() {
	Lock(key)
}

func testUnlockingKey(test *testing.T) {
    err := Unlock(key);
    if (err != nil) {
        test.Errorf("Failed to unlock a key");
    }
}

func testConcurrentLockingUnlocking(test *testing.T) {
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
}

func testUnlockingAnUnlockedLock(test *testing.T) {
    err := Unlock(key);
    if (err == nil) {
        test.Errorf("Failed to unlock a key");
    }
}

func testUnlockingMissingKey(test *testing.T) {
    err := Unlock(doesNotExistKey);
    if err == nil {
        test.Errorf("Lockmanager unlocked a key that doesn't exist");
    }


}

func testLockConcurrencyWithStaleCheck(test *testing.T) {

    testWithCondition := func(shouldPreventRemoval bool) bool {
		
        Lock(key) 
        Unlock(key) 

        val, _ := lockMap.Load(key)
        val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)))

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
			RemoveStaleLocksOnce()
		}()

		if (shouldPreventRemoval) {
			Lock(key)
			Unlock(key)
		}

		wg.Wait()

        _, exists := lockMap.Load(key)
        return exists
    }

    // Test if a lock gets past the first "if" in RemoveStaleLocksOnce
	// but not the second because it had been aquired 
    if !testWithCondition(true) {
        test.Errorf("Lock was unexpectedly removed even though it was accessed concurrently")
    }


    // Test if a lock gets past both "if's" and gets removed from the map
    if testWithCondition(false) {
        test.Errorf("Stale lock was not removed")
    }

}

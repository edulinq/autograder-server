package common

import (
    "sync"
    "testing"
    "time"

    "github.com/edulinq/autograder/config"
)

func TestLockBase(test *testing.T) {
    key1 := "testkey1";

    Lock(key1);

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
}

func TestUnlockingAnUnlockedLock(test *testing.T) {
    key1 := "testkey1";

    Lock(key1);

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }

    err = Unlock(key1);
    if (err == nil) {
        test.Fatalf("Failed to return an error when unlocking a lock that is already unlocked.");
    }
}

func TestUnlockingKeyThatDoesntExist(test *testing.T) {
    doesNotExistKey := "dne";

    err := Unlock(doesNotExistKey);
    if (err == nil) {
        test.Fatalf("Failed to return an error when unlocking a key that does not exist.");
    }
}

func TestLockUnlockDifferentKeysOneThread(test *testing.T) {
    key1 := "testkey1";
    key2 := "testkey2";
    var verifyWaitGroup sync.WaitGroup;
    var threadWaitGroup sync.WaitGroup;

    // Lock key1 for the first time.
    Lock(key1);

    threadWaitGroup.Add(1);
    verifyWaitGroup.Add(1);

    // This thread should lock key2 for the first time while key1 is already locked.
    go func() {
        defer threadWaitGroup.Done();
        Lock(key2);
        verifyWaitGroup.Done();
        err := Unlock(key2);
        if (err != nil) {
            test.Fatalf("Failed to unlock.");
        }
    }()

    // Make sure the thread locked key2 before the main thread unlocks key1.
    verifyWaitGroup.Wait();

    // Unlock key1.
    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }

    // Wait for the thread to finish unlocking key2.
    threadWaitGroup.Wait();
}

func TestLockUnlockDifferentKeysTwoThreads(test *testing.T) {
    key1 := "testkey1";
    key2 := "testkey2";
    var threadsWaitGroup sync.WaitGroup;
    var funcWaitGroup sync.WaitGroup

    threadsWaitGroup.Add(2);
    funcWaitGroup.Add(2);

    // First thread locks and unlocks key1.
    go func() {
        defer funcWaitGroup.Done();
        Lock(key1);
        threadsWaitGroup.Done();
        err := Unlock(key1);
        if (err != nil) {
            test.Fatalf("Failed to unlock.")
        }
    }()

    // Second thread locks and unlocks key2.
    go func() {
        defer funcWaitGroup.Done();
        Lock(key2);
        threadsWaitGroup.Done();
        err := Unlock(key2);
        if (err != nil) {
            test.Fatalf("Failed to Unlock")
        }
    }()
    
    // Wait until both threads have finished locking their keys so they 
    // can unlock at the same time.
    threadsWaitGroup.Wait();
    
    // Wait until both threads finish their execution.
    funcWaitGroup.Wait();
}

func TestLockingTwiceWithSameKeyConcurrently(test *testing.T) {
    key1 := "testkey1";
    var firstThreadWaitGroup sync.WaitGroup;
    var secondThreadWaitGroup sync.WaitGroup;
    var bothThreadsWaitGroup sync.WaitGroup;

    firstThreadWaitGroup.Add(1);
    secondThreadWaitGroup.Add(1);
    bothThreadsWaitGroup.Add(1);

    // First thread aquires the lock a first time.
    go func() {
        defer firstThreadWaitGroup.Done();
        Lock(key1);
    }()

    // Wait for the first thread to lock key1. 
    firstThreadWaitGroup.Wait();


    // Second thread aquires the lock a second time.
    go func() {
        defer bothThreadsWaitGroup.Done();
        secondThreadWaitGroup.Done();
        Lock(key1);
    }()

    // Wait for the second thread to try to lock key1
    // while the lock is being used by the first thread.
    secondThreadWaitGroup.Wait();
    
    // Small sleep to ensure the second thread gets to locking key1.
    time.Sleep(10 * time.Millisecond);

    // Unlock the first threads lock.
    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
    
    // Wait for the second thread to aquire the lock.
    bothThreadsWaitGroup.Wait();

    // Unlock the second threads lock.
    err = Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
}

func TestLockUnlockSameKeyConcurrently(test *testing.T) {
    key1 := "testkey1";
    var threadWaitGroup sync.WaitGroup;
    var verifyWaitGroup sync.WaitGroup;

    // Lock key1 for the first time.
    Lock(key1);

    threadWaitGroup.Add(1);
    verifyWaitGroup.Add(1);

    // This thread should wait to lock/unlock key1 a second time until it gets
    // unlocked for the first time.
    go func() {
        defer threadWaitGroup.Done();
        verifyWaitGroup.Done();
        Lock(key1);
        defer func() {
            err := Unlock(key1);
            if (err != nil) {
                test.Fatalf("Failed to unlock.");
            }
        }()
    }()

    // Make sure the thread is trying to Lock with the same key
    // before the main thread unlocks it.
    verifyWaitGroup.Wait();

    // Small sleep to ensure the thread gets to locking key1.
    time.Sleep(10 * time.Millisecond);

    // Unlock key1 for the first time.
    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }

    // Wait for the thread to finish locking and unlocking key1.
    threadWaitGroup.Wait();
}

func TestStaleLockRetention(test *testing.T) {
    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var threadWaitGroup sync.WaitGroup;
    var timestampWaitGroup sync.WaitGroup;

    Lock(key1);
    Unlock(key1);

    // Load the lock data.
    val, _ := lockMap.Load(key1); 
    // Set the timestamp for a lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));

    threadWaitGroup.Add(1);
    timestampWaitGroup.Add(1);

    // Start a goroutine to check if the lock is stale.
    go func() {
        defer threadWaitGroup.Done();
        // Wait for the lock's timestamp to be reset.
        timestampWaitGroup.Wait();
        RemoveStaleLocksOnce();
    }()

    // Reset the locks timestamp.
    Lock(key1);
    Unlock(key1);
    
    // Let the goroutine continue its execution after resetting the timestamp.
    timestampWaitGroup.Done();

    // Wait for the go routine to finish its execution.
    threadWaitGroup.Wait();
    
    // Check if the lock still exists in the lock map.
    _, exists := lockMap.Load(key1);
    // Test if the lock gets past the first check in RemoveStaleLocksOnce but not the second because it had been acquired.
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was accessed concurrently.");
    }
} 

func TestStaleLockRemoval(test *testing.T) {
    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var threadWaitGroup sync.WaitGroup;

    Lock(key1);
    Unlock(key1);

    // Load the lock data.
    val, _ := lockMap.Load(key1); 
    // Set the timestamp for a lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));

    threadWaitGroup.Add(1);

    // Start a goroutine to check if the lock is stale.
    go func() {
        defer threadWaitGroup.Done();
        RemoveStaleLocksOnce();
    }()
    
    // Wait for the goroutine to finish checking for staleness.
    threadWaitGroup.Wait();
    
    // Check if the lock still exists in the lock map.
    _, exists := lockMap.Load(key1);
    // Test if the lock gets past the first and second check in RemoveStaleLocksOnce and gets deleted.
    if (exists) {
        test.Fatalf("Failed to remove stale lock.");
    }
}

func TestStaleLockRemovalWithLockedKey(test *testing.T) {
    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var lockWaitGroup sync.WaitGroup;

    Lock(key1);

    // Load the lock data.
    val, _ := lockMap.Load(key1); 
    // Set the timestamp for a lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));
    
    lockWaitGroup.Add(1);

    // Start a goroutine to check if the lock is stale.
    go func() {
        defer lockWaitGroup.Done();
        RemoveStaleLocksOnce();
    }()
    
    // Wait for the goroutine to finish its execution.
    lockWaitGroup.Wait();

    // Check if the lock still exists in the lock map and return it.
    _, exists := lockMap.Load(key1);
    // Test if the Locked lock got removed.
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was locked.");
    }
}

package common

import (
    "sync"
    "testing"
    "time"

    "github.com/edulinq/autograder/config"
)

func Clear() {
    lockManagerMutex = sync.Mutex{};

    lockMap = sync.Map{};
}

// Lock(1) -> Unlock(1).
func TestLockBase(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";

    Lock(key1);

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
}

// Lock(1) -> Unlock(1) -> Unlock(1).
func TestUnlockingAnUnlockedLock(test *testing.T) {
    Clear();
    defer Clear();

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

// Unlock(1).
func TestUnlockingKeyThatDoesntExist(test *testing.T) {
    Clear();
    defer Clear();

    doesNotExistKey := "dne";

    err := Unlock(doesNotExistKey);
    if (err == nil) {
        test.Fatalf("Failed to return an error when unlocking a key that does not exist.");
    }
}

// Lock(1) -> %Lock(2) -> Unlock(1) -> %Unlock(2).
// % denotes a thread.
func TestThread1UnlockFirst(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    key2 := "testkey2";
    var key1UnlockBlock sync.WaitGroup;
    var key2LockedBlock sync.WaitGroup;
    var threadFinishBlock sync.WaitGroup;

    Lock(key1);

    threadFinishBlock.Add(1);
    key2LockedBlock.Add(1);
    key1UnlockBlock.Add(1);

    go func() {
        defer threadFinishBlock.Done();

        // Lock key2 while key1 is already locked.
        Lock(key2);

        key2LockedBlock.Done();

        // Wait for the main thread to unlock key1.
        key1UnlockBlock.Wait();

        err := Unlock(key2);
        if (err != nil) {
            test.Fatalf("Failed to unlock inside the thread.");
        }
    }()

    // Wait for the thread to lock key2 before unlocking key1.
    key2LockedBlock.Wait();

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }

    key1UnlockBlock.Done();

    // Wait for the thread to finish unlocking key2.
    threadFinishBlock.Wait();
}

// Lock(1) -> %Lock(2) -> %Unlock(2) -> Unlock(1).
// % denotes a thread.
func TestThread2UnlockFirst(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    key2 := "testkey2";
    var key2LockUnlockBlock sync.WaitGroup;

    Lock(key1);

    key2LockUnlockBlock.Add(1);

    go func() {
        defer key2LockUnlockBlock.Done();

        // Lock key2 while key1 is already locked.
        Lock(key2);

        err := Unlock(key2);
        if (err != nil) {
            test.Fatalf("Failed to unlock inside the thread.");
        }
    }()

    // Wait for the thread to lock and unlock key2 before unlocking key1.
    key2LockUnlockBlock.Wait();

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
}

// Lock(1) -> %Lock(1) -> Unlock(1) -> %Unlock(1).
// % denotes a thread.
func TestLockingTwiceWithSameKeyConcurrently(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    var threadLockBlock sync.WaitGroup;
    var threadFinishBlock sync.WaitGroup;

    threadLockBlock.Add(1);
    threadFinishBlock.Add(1);

    // Lock for the first time.
    Lock(key1);

    // This thread tries to lock key1 a second time but is blocked
    // until key1 gets unlocked for the first time.
    go func() {
        defer threadFinishBlock.Done();

        threadLockBlock.Done();

        Lock(key1);
        err := Unlock(key1);
        if (err != nil) {
            test.Fatalf("Failed to unlock for the second time.");
        }
    }()

    // Wait for the second thread to try to lock key1
    // while the lock is being used by the first thread.
    threadLockBlock.Wait();
    
    // Small sleep to ensure the thread tries to lock key1.
    time.Sleep(10 * time.Millisecond);

    // Unlock for the first time.
    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock for the first time.");
    }
    
    // Wait for the thread to lock and unlock key1.
    threadFinishBlock.Wait();
}

func TestStaleLockWithNonStaleLock(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Minute;
    var threadExecutionBlock sync.WaitGroup;
    var timestampBlock sync.WaitGroup;

    Lock(key1);
    Unlock(key1);

    val, _ := lockMap.Load(key1); 
    // Set the timestamp for the lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Minute)));

    threadExecutionBlock.Add(1);
    timestampBlock.Add(1);

    go func() {
        defer threadExecutionBlock.Done();

        // Wait for the lock's timestamp to be reset.
        timestampBlock.Wait();

        RemoveStaleLocksOnce();
    }()

    // Reset the locks timestamp.
    Lock(key1);
    Unlock(key1);
    
    // Let the goroutine continue its execution after resetting the timestamp.
    timestampBlock.Done();

    // Wait for the go routine to finish its execution.
    threadExecutionBlock.Wait();
    
    // Check if the lock still exists in the lock map.
    _, exists := lockMap.Load(key1);
    // Test if the non stale lock got removed.
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was accessed concurrently.");
    }
} 

func TestStaleLockWithLockedKey(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var staleCheckBlock sync.WaitGroup;

    Lock(key1);

    val, _ := lockMap.Load(key1); 
    // Set the timestamp for a lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));
    
    staleCheckBlock.Add(1);

    // Start a goroutine to check if the lock is stale.
    go func() {
        defer staleCheckBlock.Done();

        RemoveStaleLocksOnce();
    }()
    
    // Wait for the goroutine to finish its execution.
    staleCheckBlock.Wait();

    // Check if the lock still exists in the lock map and return it.
    _, exists := lockMap.Load(key1);
    // Test if the Locked lock got removed.
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was locked.");
    }
}

func TestStaleLockPassesFirstCheckButNotSecond(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var finishThreadBlock sync.WaitGroup;

    Lock(key1);
    Unlock(key1);

    val, _ := lockMap.Load(key1);
    lockData := val.(*lockData);
    // Set the timestamp for a lock to be considered stale.
    lockData.timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Minute)));

    finishThreadBlock.Add(1);

    // Lock the lockManagerMutex to to give the main thread time to reset the lock's timestamp.
    lockManagerMutex.Lock();

    // This thread will pass the first guard but wait until the lock's timestamp gets reset
    // to pass the second guard.
    go func() {
        defer finishThreadBlock.Done();

        RemoveStaleLocksOnce();
    }()

    // Small sleep to give the thread time to pass the first check in RemoveStaleLocksOnce().
    time.Sleep(10 * time.Millisecond);

    // Reset the lockdata's timestamp to simulate a lock aquiring a lock between checks in RemoveStaleLocksOnce().
    lockData.timestamp = time.Now();
    
    // Let the thread continue to the second check in RemoveStaleLocksOnce().
    lockManagerMutex.Unlock();

    // Wait until the thread finishes.
    finishThreadBlock.Wait();

    _, exists := lockMap.Load(key1);
    // Test if the lock gets past the first check but not the second in RemoveStaleLocksOnce().
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was accessed concurrently.");
    }
}

func TestStaleLockDeletion(test *testing.T) {
    Clear();
    defer Clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var staleCheckBlock sync.WaitGroup;

    Lock(key1);
    Unlock(key1);

    val, _ := lockMap.Load(key1); 
    // Set the timestamp for a lock to be considered stale.
    val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));

    staleCheckBlock.Add(1);

    go func() {
        defer staleCheckBlock.Done();

        RemoveStaleLocksOnce();
    }()
    
    // Wait for the goroutine to finish checking for staleness.
    staleCheckBlock.Wait();
    
    _, exists := lockMap.Load(key1);
    // Test if the stale lock got deleted.
    if (exists) {
        test.Fatalf("Failed to remove stale lock.");
    }
}

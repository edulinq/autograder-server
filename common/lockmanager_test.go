package common

// Function level comments: 
// % denotes a thread off the main thread.
// ! denotes an error from doing an operation.
// Ex. Lock(1) -> %!Unlock(2) - The main thread locks key1 then a thread generates an error unlocking key2.

import (
    "sync"
    "testing"
    "time"
    
    "github.com/edulinq/autograder/config"
)

func clear() {
    lockManagerMutex = sync.Mutex{};
    lockMap = sync.Map{};
}

// Lock(1) -> Unlock(1).
func TestLockBase(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";

    Lock(key1);

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock.");
    }
}

// Lock(1) -> Unlock(1) -> !Unlock(1).
func TestUnlockingAnUnlockedLock(test *testing.T) {
    clear();
    defer clear();

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
    clear();
    defer clear();

    doesNotExistKey := "dne";

    err := Unlock(doesNotExistKey);
    if (err == nil) {
        test.Fatalf("Failed to return an error when unlocking a key that does not exist.");
    }
}

// Lock(1) -> %Lock(2) -> Unlock(1) -> %Unlock(2).
func TestMainThreadUnlocksFirst(test *testing.T) {
    clear();
    defer clear();

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
            test.Fatalf("Failed to unlock key2 in the child thread.");
        }
    }();

    // Wait for the thread to lock key2 before unlocking key1.
    key2LockedBlock.Wait();

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock key1 in the main thread.");
    }

    key1UnlockBlock.Done();

    // Wait for the thread to finish unlocking key2.
    threadFinishBlock.Wait();
}

// Lock(1) -> %Lock(2) -> %Unlock(2) -> Unlock(1).
func TestChildThreadUnlocksFirst(test *testing.T) {
    clear();
    defer clear();

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
            test.Fatalf("Failed to unlock key2 in the child thread.");
        }
    }();

    // Wait for the thread to lock and unlock key2 before unlocking key1.
    key2LockUnlockBlock.Wait();

    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock key1 in the main thread.");
    }
}

// Lock(1) -> %Lock(1) -> Unlock(1) -> %Unlock(1).
func TestMainThreadUnlockFirst(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";
    var threadLockBlock sync.WaitGroup;
    var threadFinishBlock sync.WaitGroup;

    threadLockBlock.Add(1);
    threadFinishBlock.Add(1);
    unlockInChildThreadFirst := make(chan struct{}, 1);

    // Lock for the first time.
    Lock(key1);

    // This thread tries to lock key1 a second time but is blocked
    // until key1 gets unlocked for the first time.
    go func() {
        defer threadFinishBlock.Done();

        threadLockBlock.Done();

        Lock(key1);

        // Signal the child thread locked key1.
        unlockInChildThreadFirst <- struct{}{};

        err := Unlock(key1);
        if (err != nil) {
            test.Fatalf("Failed to unlock for the second time in the child thread.");
        }
    }();

    // Wait for the second thread to try to lock key1
    // while the lock is being used by the first thread.
    threadLockBlock.Wait();
    
    // Small sleep to ensure the thread tries to lock key1.
    time.Sleep(10 * time.Millisecond);

    // Check if the child thread unlocked key1 before the main thread did.
    if (len(unlockInChildThreadFirst) > 0) {
        test.Fatalf("Failed to ensure the main thread unlocked key1 before the child thread unlocked it.");
    }

    // Unlock for the first time.
    err := Unlock(key1);
    if (err != nil) {
        test.Fatalf("Failed to unlock for the first time in the main thread.");
    }
    
    // Wait for the thread to lock and unlock key1.
    threadFinishBlock.Wait();
}

// Lock(1) -> Unlock(1) -> %StaleCheck.
func TestStaleRetentionWithNonStaleLock(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";
    var threadExecutionBlock sync.WaitGroup;

    // Add key1 to the lockmap and have it unlocked.
    Lock(key1);
    Unlock(key1);

    threadExecutionBlock.Add(1);

    go func() {
        defer threadExecutionBlock.Done();

        RemoveStaleLocksOnce();
    }();

    // Wait for the thread to finish checking for staleness.
    threadExecutionBlock.Wait();
    
    // Check if the stale lock got removed.
    _, exists := lockMap.Load(key1);
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was accessed concurrently.");
    }
} 

// Lock(1) -> %Stale Check.
func TestStaleLockRetentionWithLockedKey(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var staleCheckBlock sync.WaitGroup;

    // Add key1 to the lockmap and have it locked.
    Lock(key1);

    // Load the lockmap and set the timestamp for the lock to be considered stale.
    val, _ := lockMap.Load(key1); 
    val.(*lockData).timestamp = time.Now().Add(-2 * (staleDuration));
    
    staleCheckBlock.Add(1);

    // Start a goroutine to check if the lock is stale.
    go func() {
        defer staleCheckBlock.Done();

        RemoveStaleLocksOnce();
    }();
    
    // Wait for the thread to finish checking for staleness.
    staleCheckBlock.Wait();

    // Check if the locked stale lock got removed. 
    _, exists := lockMap.Load(key1);
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was locked.");
    }
}

// Lock(1) -> Unlock(1) -> Make Stale -> %Stale Check first check -> Update timestamp -> %Stale Check second check. 
func TestLockRetentionWithMidCheckActivity(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var finishThreadBlock sync.WaitGroup;
    var threadStartBlock sync.WaitGroup;

    // Add key1 to the lockmap and have it unlocked.
    Lock(key1);
    Unlock(key1);

    // Load the lockmap and set the timestamp for the lock to be considered stale.
    val, _ := lockMap.Load(key1);
    lockData := val.(*lockData);
    lockData.timestamp = time.Now().Add(-2 * (staleDuration));

    finishThreadBlock.Add(1);
    threadStartBlock.Add(1);

    // Lock the lockManagerMutex to to give the main thread time to reset the lock's timestamp.
    lockManagerMutex.Lock();

    // This thread will pass the first check but wait until the lock's timestamp gets reset
    // to pass the second check.
    go func() {
        defer finishThreadBlock.Done();

        threadStartBlock.Done();

        RemoveStaleLocksOnce();
    }();

    // Wait for the thread to start.
    threadStartBlock.Wait();

    // Small sleep to give the thread time to pass the first stale check in RemoveStaleLocksOnce().
    time.Sleep(10 * time.Millisecond);

    // Reset the lockdata's timestamp to simulate a lock aquiring a lock between checks in RemoveStaleLocksOnce().
    lockData.timestamp = time.Now();
    
    // Let the thread continue to the second stale check in RemoveStaleLocksOnce().
    lockManagerMutex.Unlock();

    // Wait for the thread to finish checking for staleness.
    finishThreadBlock.Wait();

    // Check if the stale lock changed to un-stale mid check got removed.
    _, exists := lockMap.Load(key1);
    if (!exists) {
        test.Fatalf("Failed to retain lock even though it was accessed concurrently.");
    }
}

// Lock(1) -> Unlock(1) -> Make Stale -> %Stale Check
func TestStaleLockDeletion(test *testing.T) {
    clear();
    defer clear();

    key1 := "testkey1";
    staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    var finishThreadBlock sync.WaitGroup;

    // Add key1 to the lockmap and have it unlocked.
    Lock(key1);
    Unlock(key1);

    // Load the lockmap and set the timestamp for the lock to be considered stale.
    val, _ := lockMap.Load(key1); 
    val.(*lockData).timestamp = time.Now().Add(-2 * (staleDuration));

    finishThreadBlock.Add(1);

    go func() {
        defer finishThreadBlock.Done();

        RemoveStaleLocksOnce();
    }();
    
    // Wait for the thread to finish checking for staleness.
    finishThreadBlock.Wait();
    
    // Check if the stale lock got deleted.
    _, exists := lockMap.Load(key1);
    if (exists) {
        test.Fatalf("Failed to remove stale lock.");
    }
}

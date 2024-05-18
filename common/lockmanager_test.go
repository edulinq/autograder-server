package common

import (
    "sync"
    "testing"
    "time"

    "github.com/edulinq/autograder/config"
)

var (
    key1 = "testkey1";
    key2 = "testkey2";
    doesNotExistKey = "dne";
    staleDuration = time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second;
    lockWaitGroup sync.WaitGroup;
)

func TestLockBase(t *testing.T) {
    Lock(key1);

    err := Unlock(key1);
    if err != nil {
        t.Errorf("Failed to unlock a key.");
    }
}

func TestUnlockingAnUnlockedLock(t *testing.T) {
    Lock(key1);

    err := Unlock(key1);
    if err != nil {
        t.Errorf("Failed to unlock a key.");
    }

    err = Unlock(key1);
    if err == nil {
        t.Errorf("Unlocking a lock that is already unlocked did not return an error.");
    }
}

func TestUnlockingKeyThatIsntLocked(t *testing.T) {
    err := Unlock(doesNotExistKey);
    if err == nil {
        t.Errorf("Lock manager unlocked a key that wasn't locked.");
    }
}

func TestConcurrentLockingUnlocking(t *testing.T) {
    // Lock key1 and key2 for the first time.
    Lock(key1);
    Lock(key2);

    lockWaitGroup.Add(2);

    // This goroutine should wait to lock/unlock key1 a second time until it gets
    // unlocked for the first time.
    go func() {
        defer lockWaitGroup.Done();
        Lock(key1);
        defer func() {
            err := Unlock(key1);
            if (err != nil) {
                t.Errorf("Failed to unlock a key.");
            }
        }()
    }()

    // This goroutine should wait to lock/unlock key2 a second time until it gets
    // unlocked for the first time.
    go func() {
        defer lockWaitGroup.Done();
        Lock(key2);
        defer func() {
            err := Unlock(key2);
            if (err != nil) {
                t.Errorf("Failed to unlock a key.");
            }
        }()
    }()
    
    // Unlock key1 & key2 for the first time.
    err := Unlock(key1);
    if err != nil {
        t.Errorf("Failed to unlock a key.");
    }
    err = Unlock(key2);
    if err != nil {
        t.Errorf("Failed to unlock a key.");
    }

    lockWaitGroup.Wait();
}

func TestLockConcurrencyWithStaleCheck(t *testing.T) {
    testWithCondition := func(shouldPreventRemoval bool) bool {
        Lock(key1);
        Unlock(key1);

        val, _ := lockMap.Load(key1);
        val.(*lockData).timestamp = time.Now().Add(-1 * (staleDuration + (1 * time.Second)));

        lockWaitGroup.Add(1);
        go func() {
            defer lockWaitGroup.Done();
            time.Sleep(100 * time.Millisecond);
            RemoveStaleLocksOnce();
        }()

        if shouldPreventRemoval {
            Lock(key1);
            Unlock(key1);
        }

        lockWaitGroup.Wait();

        _, exists := lockMap.Load(key1);
        return exists;
    }

    // Test if a lock gets past the first "if" in RemoveStaleLocksOnce but not the second because it had been acquired.
    if !testWithCondition(true) {
        t.Errorf("Lock was unexpectedly removed even though it was accessed concurrently.");
    }

    // Test if a lock gets past both "if's" and gets removed from the map.
    if testWithCondition(false) {
        t.Errorf("Stale lock was not removed.");
    }
}

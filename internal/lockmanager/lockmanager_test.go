package lockmanager

// Tests in this file will deal with locking, unlocking, and checking for stale locks (sometimes on multiple threads).
// Since this can quickly get confusing, we will have comments for each test using the following syntax:
// "%" denotes a thread off the main thread.
// "!" denotes that the associated operation should produce an error.
// Ex. Lock(key1) -> %!Unlock(key2) - The main thread locks key1 then a thread generates an error unlocking key2.

import (
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/config"
)

// Lock(key1) -> Unlock(key1).
func TestLockBase(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"

	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the lock.")
	}
}

// Lock(key1) -> Unlock(key1) -> !Unlock(key1).
func TestUnlockingAnUnlockedLock(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"

	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the lock.")
	}

	err = Unlock(key1)
	if err == nil {
		test.Fatalf("Failed to return an error when unlocking a lock that is already unlocked.")
	}
}

// !Unlock(key1).
func TestUnlockingKeyThatDoesntExist(test *testing.T) {
	clear()
	defer clear()

	doesNotExistKey := "dne"

	err := Unlock(doesNotExistKey)
	if err == nil {
		test.Fatalf("Failed to return an error when unlocking a key that does not exist.")
	}
}

// Lock(key1) -> %Lock(key1) -> Unlock(key1) -> %Unlock(key1).
func TestMainThreadUnlocksFirstWithOneLock(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	var threadLockBlock sync.WaitGroup
	var threadFinishBlock sync.WaitGroup
	unlockInChildThreadFirst := make(chan bool, 1)

	// Lock for the first time.
	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	threadLockBlock.Add(1)
	threadFinishBlock.Add(1)

	// This thread tries to lock key1 a second time but is blocked
	// until key1 gets unlocked for the first time.
	go func() {
		defer threadFinishBlock.Done()

		threadLockBlock.Done()

		noWait = Lock(key1)
		if noWait {
			test.Fatalf("Unexpected wait for the lock (in thread): '%v'.", noWait)
		}

		// Signal the thread locked key1.
		unlockInChildThreadFirst <- true

		err := Unlock(key1)
		if err != nil {
			test.Fatalf("Failed to unlock the lock for the second time in the child thread.")
		}
	}()

	// Wait for the second thread to try to lock key1
	// while the lock is being used by the first thread.
	threadLockBlock.Wait()

	// Small sleep to ensure the thread tries to lock key1.
	time.Sleep(10 * time.Millisecond)

	// Check if the thread unlocked key1 before the main thread did.
	if len(unlockInChildThreadFirst) > 0 {
		test.Fatalf("Failed to ensure the main thread unlocked key1 before the child thread unlocked it.")
	}

	// Unlock for the first time.
	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the lock for the first time in the main thread.")
	}

	// Wait for the thread to lock and unlock key1.
	threadFinishBlock.Wait()

	// Check if the thread was able to lock key1 after the main thread unlocked key1.
	if len(unlockInChildThreadFirst) == 0 {
		test.Fatalf("Failed to Lock in the child thread.")
	}
}

// Lock(key1) -> %Lock(key2) -> Unlock(key1) -> %Unlock(key2).
func TestMainThreadUnlocksFirstWithTwoLocks(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	key2 := "testkey2"
	var key1UnlockBlock sync.WaitGroup
	var key2LockedBlock sync.WaitGroup
	var threadFinishBlock sync.WaitGroup

	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for lock 1: '%v'.", noWait)
	}

	threadFinishBlock.Add(1)
	key2LockedBlock.Add(1)
	key1UnlockBlock.Add(1)

	go func() {
		defer threadFinishBlock.Done()

		// Lock key2 while key1 is already locked.
		noWait = Lock(key2)
		if !noWait {
			test.Fatalf("Unexpected wait for lock 2: '%v'.", noWait)
		}

		key2LockedBlock.Done()

		// Wait for the main thread to unlock key1.
		key1UnlockBlock.Wait()

		err := Unlock(key2)
		if err != nil {
			test.Fatalf("Failed to unlock lock key2 in the child thread.")
		}
	}()

	// Wait for the thread to lock key2 before unlocking key1.
	key2LockedBlock.Wait()

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock key1 in the main thread.")
	}

	key1UnlockBlock.Done()

	// Wait for the thread to finish unlocking key2.
	threadFinishBlock.Wait()
}

// Lock(key1) -> %Lock(key2) -> %Unlock(key2) -> Unlock(key1).
func TestChildThreadUnlocksFirstWithTwoLocks(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	key2 := "testkey2"
	var key2LockUnlockBlock sync.WaitGroup

	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for lock 1: '%v'.", noWait)
	}

	key2LockUnlockBlock.Add(1)

	go func() {
		defer key2LockUnlockBlock.Done()

		// Lock key2 while key1 is already locked.
		noWait = Lock(key2)
		if !noWait {
			test.Fatalf("Unexpected wait for lock 2: '%v'.", noWait)
		}

		err := Unlock(key2)
		if err != nil {
			test.Fatalf("Failed to unlock key2 in the child thread.")
		}
	}()

	// Wait for the thread to lock and unlock key2 before unlocking key1.
	key2LockUnlockBlock.Wait()

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock key1 in the main thread.")
	}
}

// Lock(key1) -> Unlock(key1) -> %StaleCheck.
func TestStaleRetentionWithNonStaleLock(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	var threadExecutionBlock sync.WaitGroup

	// Add key1 to the lockmap and have it unlocked.
	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	Unlock(key1)

	threadExecutionBlock.Add(1)

	go func() {
		defer threadExecutionBlock.Done()

		RemoveStaleLocksOnce()
	}()

	// Wait for the thread to finish checking for staleness.
	threadExecutionBlock.Wait()

	// Check if the non-stale lock got removed.
	_, exists := lockMap.Load(key1)
	if !exists {
		test.Fatalf("Failed to retain lock even though it was accessed concurrently.")
	}
}

// Lock(key1) -> MakeTimestampStale(key1) -> %StaleCheck.
func TestStaleLockRetentionWithLockedKey(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second
	var staleCheckBlock sync.WaitGroup

	// Add key1 to the lockmap and have it locked.
	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	// Load the lockmap and set the timestamp for the lock to be considered stale.
	val, _ := lockMap.Load(key1)
	val.(*lockData).timestamp = time.Now().Add(-2 * (staleDuration))

	staleCheckBlock.Add(1)

	// Start a goroutine to check if the lock is stale.
	go func() {
		defer staleCheckBlock.Done()

		RemoveStaleLocksOnce()
	}()

	// Wait for the thread to finish checking for staleness.
	staleCheckBlock.Wait()

	// Check if the locked stale lock got removed.
	_, exists := lockMap.Load(key1)
	if !exists {
		test.Fatalf("Failed to retain lock even though it was locked.")
	}

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock key1 after checking for staleness.")
	}
}

// Lock(key1) -> Unlock(key1) -> MakeTimestampStale(key1) -> %StaleCheck(First Part) -> MakeTimestampNotStale(key1) -> %StaleCheck(Last Part).
func TestLockRetentionWithMidCheckActivity(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second
	var finishThreadBlock sync.WaitGroup
	var threadStartBlock sync.WaitGroup

	// Add key1 to the lockmap and have it unlocked.
	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	Unlock(key1)

	// Load the lockmap and set the timestamp for the lock to be considered stale.
	val, _ := lockMap.Load(key1)
	lockData := val.(*lockData)
	lockData.timestamp = time.Now().Add(-2 * (staleDuration))

	// Lock the lockManagerMutex to to give the main thread time to reset the lock's timestamp.
	lockManagerMutex.Lock()

	finishThreadBlock.Add(1)
	threadStartBlock.Add(1)

	// This thread will pass the first part but wait until the lock's timestamp gets reset
	// to pass the second part of RemoveStaleLocksOnce().
	go func() {
		defer finishThreadBlock.Done()

		threadStartBlock.Done()

		RemoveStaleLocksOnce()
	}()

	// Wait for the thread to start.
	threadStartBlock.Wait()

	// Small sleep to give the thread time to pass the first part in RemoveStaleLocksOnce().
	time.Sleep(10 * time.Millisecond)

	// Reset the lockdata's timestamp to simulate a lock aquiring a lock between checks in RemoveStaleLocksOnce().
	lockData.timestamp = time.Now()

	// Let the thread continue to the second part in RemoveStaleLocksOnce().
	lockManagerMutex.Unlock()

	// Wait for the thread to finish checking for staleness.
	finishThreadBlock.Wait()

	// Check if the stale lock changed to not stale mid check got removed.
	_, exists := lockMap.Load(key1)
	if !exists {
		test.Fatalf("Failed to retain lock even though it was accessed concurrently.")
	}
}

// Lock(key1) -> Unlock(key1) -> MakeTimestampStale(key1) -> %StaleCheck.
func TestStaleLockDeletion(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second
	var finishThreadBlock sync.WaitGroup

	// Add key1 to the lockmap and have it unlocked.
	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	Unlock(key1)

	// Load the lockmap and set the timestamp for the lock to be considered stale.
	val, _ := lockMap.Load(key1)
	val.(*lockData).timestamp = time.Now().Add(-2 * (staleDuration))

	finishThreadBlock.Add(1)

	go func() {
		defer finishThreadBlock.Done()

		RemoveStaleLocksOnce()
	}()

	// Wait for the thread to finish checking for staleness.
	finishThreadBlock.Wait()

	// Check if the stale lock got deleted.
	_, exists := lockMap.Load(key1)
	if exists {
		test.Fatalf("Failed to remove stale lock.")
	}
}

// ReadLock(key1) -> ReadUnlock(key1).
func TestReadLockBase(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"

	ReadLock(key1)

	err := ReadUnlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the read lock.")
	}
}

// ReadLock(key1) -> %ReadLock(key1) -> %ReadUnlock(key1) -> ReadUnlock(key1).
func TestSimultaneousReadLocks(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	var finishThreadBlock sync.WaitGroup
	var readLockChildThreadBlock sync.WaitGroup
	readLockInChildThread := make(chan bool, 1)

	ReadLock(key1)

	finishThreadBlock.Add(1)
	readLockChildThreadBlock.Add(1)

	go func() {
		defer finishThreadBlock.Done()

		ReadLock(key1)

		// Signal that key1 read locked.
		readLockInChildThread <- true
		readLockChildThreadBlock.Done()

		err := ReadUnlock(key1)
		if err != nil {
			test.Fatalf("Failed to unlock the read lock for key2 in the child thread.")
		}
	}()

	readLockChildThreadBlock.Wait()

	// Check if the thread was able to read lock key1 a second time.
	if len(readLockInChildThread) == 0 {
		test.Fatalf("Failed to ensure the read lock doesn't block after another read lock.")
	}

	// Wait for the thread to finish read locking and read unlocking.
	finishThreadBlock.Wait()

	err := ReadUnlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the read lock for key1 in the main thread.")
	}
}

// ReadLock(key1) -> %Lock(key1) -> ReadUnlock(key1) -> %Unlock(key1)
func TestWriteLockBlock(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	var finishThreadBlock sync.WaitGroup
	var writeLockBlock sync.WaitGroup
	lockInThread := make(chan bool, 1)

	ReadLock(key1)

	finishThreadBlock.Add(1)
	writeLockBlock.Add(1)

	go func() {
		defer finishThreadBlock.Done()

		writeLockBlock.Done()

		noWait := Lock(key1)
		if noWait {
			test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
		}

		// Signal key1 got locked.
		lockInThread <- true

		err := Unlock(key1)
		if err != nil {
			test.Fatalf("Failed to unlock the write lock for key1 in the child thread.")
		}
	}()

	// Wait until the thread is about to lock key1.
	writeLockBlock.Wait()

	// Small sleep to ensure the thread tries to lock key1.
	time.Sleep(10 * time.Millisecond)

	// Check if the thread locked key1 while it was already read locked.
	if len(lockInThread) > 0 {
		test.Fatalf("Failed to block the write lock from locking after read locking key1.")
	}

	err := ReadUnlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the read lock for key1 in the main thread.")
	}

	// Wait for the thread to finish locking and unlocking key1.
	finishThreadBlock.Wait()

	// Check if the thread locked key1.
	if len(lockInThread) == 0 {
		test.Fatalf("Failed to lock key1 in the child thread.")
	}
}

// Lock(1) -> %ReadLock(1) -> Unlock(1) -> %ReadUnlock(1).
func TestReadLockBlock(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	readLockInChildThread := make(chan bool, 1)
	var finishThreadBlock sync.WaitGroup
	var readLockChildThreadBlock sync.WaitGroup

	noWait := Lock(key1)
	if !noWait {
		test.Fatalf("Unexpected wait for the lock: '%v'.", noWait)
	}

	finishThreadBlock.Add(1)
	readLockChildThreadBlock.Add(1)

	go func() {
		defer finishThreadBlock.Done()

		readLockChildThreadBlock.Done()

		ReadLock(key1)

		readLockInChildThread <- true

		err := ReadUnlock(key1)
		if err != nil {
			test.Fatalf("Failed to unlock the read lock in the child thread.")
		}
	}()

	// Wait until the thread is about to read lock key1.
	readLockChildThreadBlock.Wait()

	// Small sleep to ensure the thread tries to read lock key1.
	time.Sleep(10 * time.Millisecond)

	// Check if the thread read locked key1 while it was already locked.
	if len(readLockInChildThread) > 0 {
		test.Fatalf("Failed to block the read lock from read-locking key1 after locking key1.")
	}

	err := Unlock(key1)
	if err != nil {
		test.Fatalf("Failed to unlock the lock in the main thread.")
	}

	// Wait for the thread to finish read locking and read unlocking.
	finishThreadBlock.Wait()

	// Check if the thread read locked key1.
	if len(readLockInChildThread) == 0 {
		test.Fatalf("Failed to read lock key1 in the child thread.")
	}
}

// Lock(key1) -> Unlock(key1) -> MakeTimestampStale(key1) -> StaleCheck -> AssertPerKeyMutexFree.
//
// Regression test for: RemoveStaleLocksOnce() calling TryLock() on the per-key mutex
// to claim exclusive access before deleting the map entry, but never calling Unlock()
// afterward. This left the per-key mutex permanently acquired on the deleted lockData.
//
// Any goroutine that obtained a pointer to that lockData before the deletion (e.g., a
// grading goroutine that had already passed its lockMap.LoadOrStore call but had not yet
// reached lock.mutex.Lock()) would then block forever — a deadlock requiring a server restart.
//
// This test is deterministic and does not require -race: it directly inspects the
// per-key mutex on the deleted entry via TryLock(). Without the fix, TryLock() returns
// false because cleanup holds the mutex indefinitely. With the fix, cleanup calls
// Unlock() before returning, so TryLock() succeeds.
func TestCleanupReleasesPerKeyMutexAfterDeletion(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second

	Lock(key1)
	Unlock(key1)

	// Make the entry stale.
	val, _ := lockMap.Load(key1)
	entry := val.(*lockData)
	entry.timestamp = time.Now().Add(-2 * staleDuration)

	RemoveStaleLocksOnce()

	// Confirm the stale entry was deleted.
	_, exists := lockMap.Load(key1)
	if exists {
		test.Fatalf("Stale lock was not removed.")
	}

	// The per-key mutex on the deleted entry must be free after cleanup.
	// Without the fix, RemoveStaleLocksOnce() called TryLock() but never Unlock(),
	// permanently holding the mutex. TryLock() here would then return false.
	if !entry.mutex.TryLock() {
		test.Fatalf("Per-key mutex was permanently locked by stale lock cleanup (missing Unlock after TryLock).")
	}
	entry.mutex.Unlock()
}

// (100x) Lock(key1) -> Unlock(key1) -> MakeTimestampStale(key1) -> [%Lock(key1) || %StaleCheck].
//
// Regression test for: lock() writing lockCount and timestamp after releasing
// lockManagerMutex (i.e., without holding any lock), while RemoveStaleLocksOnce()
// read those same fields also without holding any lock. These were concurrent
// unsynchronized accesses to the same memory — a data race per Go's memory model.
//
// Run this test with -race. Without the fix, the race detector reports a DATA RACE
// between lock()'s write of lockCount/timestamp and RemoveStaleLocksOnce()'s read
// of those fields. With the fix, both accesses are guarded by lockManagerMutex and
// the race detector reports nothing.
//
// Note on the unchecked Unlock() error: cleanup may delete the map entry after
// lock() has acquired the per-key mutex but before Unlock() runs. In that case
// Unlock() returns "key not found". This is an expected outcome of the concurrent
// execution and is not a test failure — the point of this test is the absence of
// a data race, not the success of every Unlock() call.
func TestLockAndStaleLockCleanupConcurrentlyNoDataRace(test *testing.T) {
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second
	key1 := "testkey1"

	const iterations = 100

	for i := 0; i < iterations; i++ {
		clear()

		Lock(key1)
		Unlock(key1)

		val, _ := lockMap.Load(key1)
		val.(*lockData).timestamp = time.Now().Add(-2 * staleDuration)

		ready := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: lock() writes lockCount and timestamp.
		// Before the fix these writes occurred without lockManagerMutex.
		go func() {
			defer wg.Done()
			<-ready
			Lock(key1)
			// Error intentionally ignored: see note above about "key not found".
			Unlock(key1)
		}()

		// Goroutine 2: RemoveStaleLocksOnce() reads lockCount and timestamp.
		// Before the fix these reads occurred without lockManagerMutex.
		go func() {
			defer wg.Done()
			<-ready
			RemoveStaleLocksOnce()
		}()

		close(ready)
		wg.Wait()
	}
}

func clear() {
	lockManagerMutex = sync.Mutex{}
	lockMap = sync.Map{}
}

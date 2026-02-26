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
// Verifies that RemoveStaleLocksOnce() releases the per-key mutex after acquiring it
// via TryLock(). Without the fix, TryLock() was called to claim exclusive access
// before deleting the map entry, but Unlock() was never called, leaving the per-key
// mutex permanently held on the deleted lockData.
//
// Does not require -race: directly inspects the per-key mutex on the deleted entry
// via TryLock(). If the mutex is still held, TryLock() returns false.
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

// Lock(key1) -> Unlock(key1) -> MakeTimestampStale(key1) -> [%Lock(key1) || %StaleCheck].
//
// Verifies there is no data race between lock() writing lockCount/timestamp and
// RemoveStaleLocksOnce() reading those fields. Run with -race.
//
// lockManagerMutex is held externally before both goroutines start, forcing them
// to queue on it before doing any work. Goroutine 1 is started first so it enters
// the wait queue before goroutine 2. After sleeping >1ms (past Go's mutex starvation
// threshold), the queue is served FIFO, so goroutine 1 is guaranteed to acquire
// lockManagerMutex first when we release it:
//
//   - Goroutine 1 (lock()): acquires lockManagerMutex, releases it, acquires the
//     per-key mutex, then writes lockCount/timestamp — without lockManagerMutex in
//     the old code. Its vector clock never sees goroutine 2's epoch.
//   - Goroutine 2 (RemoveStaleLocksOnce()): reads lockCount/timestamp at the first
//     check without lockManagerMutex (old code), then acquires lockManagerMutex and
//     finds TryLock() fails because goroutine 1 holds the per-key mutex. Goroutine 2
//     returns without deleting the entry.
//
// No happens-before exists between goroutine 2's reads and goroutine 1's write, so
// the race detector reports a DATA RACE. With the fix, both accesses are guarded by
// lockManagerMutex — no race is reported.
func TestLockAndStaleLockCleanupConcurrentlyNoDataRace(test *testing.T) {
	clear()
	defer clear()

	key1 := "testkey1"
	staleDuration := time.Duration(config.STALELOCK_DURATION_SECS.Get()) * time.Second

	Lock(key1)
	Unlock(key1)

	val, _ := lockMap.Load(key1)
	val.(*lockData).timestamp = time.Now().Add(-2 * staleDuration)

	// Hold lockManagerMutex so both goroutines block at their first lockManagerMutex
	// acquisition before doing any racy work.
	lockManagerMutex.Lock()

	var allDone sync.WaitGroup
	allDone.Add(2)

	// Goroutine 1: lock() blocks on lockManagerMutex at function entry. Started first
	// so it enters the wait queue before goroutine 2.
	go func() {
		defer allDone.Done()
		Lock(key1)
		Unlock(key1)
	}()

	// Small delay so goroutine 1 is in the lockManagerMutex wait queue before goroutine 2 starts.
	time.Sleep(1 * time.Millisecond)

	// Goroutine 2: in the old code, RemoveStaleLocksOnce() reads lockCount/timestamp
	// at the first check without lockManagerMutex, then blocks on lockManagerMutex.
	// In the fixed code, it blocks on lockManagerMutex immediately.
	go func() {
		defer allDone.Done()
		RemoveStaleLocksOnce()
	}()

	// Sleep well past Go's 1ms mutex starvation threshold so the wait queue switches
	// to FIFO. Goroutine 1 (the earlier waiter) will acquire lockManagerMutex first.
	time.Sleep(10 * time.Millisecond)

	// Release lockManagerMutex. Goroutine 1 wakes first (FIFO starvation mode):
	// it acquires lockManagerMutex, releases it, then acquires the per-key mutex
	// and writes lockCount/timestamp without lockManagerMutex (old code). Goroutine 2
	// wakes next, acquires lockManagerMutex, and finds TryLock() fails (goroutine 1
	// holds the per-key mutex), so it returns without deleting the entry.
	// Goroutine 1's write clock never includes goroutine 2's epoch — no happens-before
	// exists between the read and the write — the race detector catches the race.
	lockManagerMutex.Unlock()

	allDone.Wait()
}

func clear() {
	lockManagerMutex = sync.Mutex{}
	lockMap = sync.Map{}
}

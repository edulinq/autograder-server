package common

import (
	"fmt"
	"testing"
	"time"
)


func TestLockAcquisition(t *testing.T) {
	lm := NewLockManager();
	key := "testkey"
	key2 := "testkey2"
	lm.Lock(key)
	lm.Lock(key2)
	fmt.Println("Main Thread!")

	go func() {
		fmt.Println("Go routine started")
		lm.Lock(key);
		defer lm.Unlock(key)

		fmt.Println("Go routine done")
		
	}()

	go func() {
		fmt.Println("2nd go routine started")
		lm.Lock(key2);
		defer lm.Unlock(key2);
		
		fmt.Println("2nd go routine done")
	}()

	fmt.Println("Main Thread Almost Done!")
	lm.Unlock(key)
	lm.Unlock(key2)

	time.Sleep(5 * time.Second);
	lm.Lock(key);
	fmt.Println("Main Thread Done")

	// fmt.Println("Got past the first lock from removeStaleLocks go routine")

	// go func() {
	// 	fmt.Println("1st Go Routine Started")

	// 	startTime := time.Now();
	// 	lm.Lock(key)
	// 	lockWaitDuration := time.Since(startTime);
	// 	fmt.Printf("Waited %v long to aquire the lock in the 1st go routine\n", lockWaitDuration);

	// 	go func() {
	// 		fmt.Println("2nd Go Routine Started")
	// 		startTime := time.Now();
	// 		lm.Lock(key);
	// 		lockWaitDuration := time.Since(startTime);
	// 		fmt.Printf("Waited %v long to aquire the lock in the 2nd go routine\n", lockWaitDuration);
	// 		defer lm.Unlock(key);
	// 		fmt.Println("2nd Go Routine Done")
	// 	}()

		
	// 	defer lm.Unlock(key)
	// 	fmt.Println("1st Go Routine Done")
	// 	time.Sleep(3 * time.Second)
	// }()

	// fmt.Println("Main Thread Almost Done")
	// time.Sleep(1 * time.Second)
	// lm.Unlock(key)

	// time.Sleep(4 * time.Second)

	// fmt.Println("Main Thread Done!")

}
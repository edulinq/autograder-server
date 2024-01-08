package common

import (
	"fmt"
	"testing"
	"time"
)


func TestLockAcquisition(t *testing.T) {
	
	key := "testkey"
	key2 := "testkey2"
	// fmt.Println("main")
	Lock(key);
	
	Lock(key2);
	fmt.Println("Main Thread!")

	go func() {
		fmt.Println("Go routine started")
		Lock(key);
		defer Unlock(key)

		fmt.Println("Go routine done")
		
	}()

	time.Sleep(1 * time.Second)

	go func() {
		fmt.Println("2nd go routine started")
		Lock(key2);
		defer Unlock(key2);
		
		fmt.Println("2nd go routine done")
	}()

	fmt.Println("Main Thread Almost Done!")
	Unlock(key)
	fmt.Println("got here")
	Unlock(key2)

	time.Sleep(2 * time.Second);
	Lock(key);
	fmt.Println("Main Thread Done")
}
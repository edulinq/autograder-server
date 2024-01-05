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
}
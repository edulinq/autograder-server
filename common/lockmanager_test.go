package common

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	key = "testkey"
	key2 = "testkey2"
)


func TestLock(t *testing.T) {
	var wg sync.WaitGroup

	Lock(key)
	Unlock(key)
	time.Sleep(4 * time.Second) // test remove stale lock

	Lock(key)
	Lock(key2)
	fmt.Println("Main Thread!")

	go func() {
		wg.Add(1)
		defer wg.Done()
		fmt.Println("Go routine started")
		Lock(key);
		time.Sleep(time.Second)
		defer Unlock(key)
		go func() {
			wg.Add(1)
			defer wg.Done()
			fmt.Println("Another go routine started")
			Lock(key2)
			defer Unlock(key2)
			fmt.Println("Another go routine done")
		}()
		fmt.Println("Go routine done")
		
	}()

	time.Sleep(5000 * time.Millisecond)

	go func() {
		wg.Add(1)
		defer wg.Done()
		fmt.Println("2nd go routine started")
		Lock(key2);
		defer Unlock(key2);
		
		fmt.Println("2nd go routine done")
	}()



	fmt.Println("Main Thread Almost Done!")
	Unlock(key)
	fmt.Println("got here")
	Unlock(key2)
	
	wg.Wait()

	fmt.Println("Main Thread Done")
}
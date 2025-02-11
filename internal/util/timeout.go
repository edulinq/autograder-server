package util

import (
	"context"
	"time"
)

func RunWithTimeout(timeoutMS int64, targetFunc func()) bool {
	return RunWithTimeoutFull(0, timeoutMS, context.Background(), func(_ context.Context) {
		targetFunc()
	})
}

// Run the given function with a timeout.
// When the soft timeout is reached, the context passed to the function will complete.
// This soft timeout context will used the passed in context as their parent.
// When the hard timeout is reached, this function will return with a false value
// (the called routin will continue to run if it does not observe the passed context).
// If the function completes before the hard timeout, then true will be returned.
func RunWithTimeoutFull(softTimeoutMS int64, hardTimeoutMS int64, ctx context.Context, targetFunc func(context.Context)) bool {
	successChan := make(chan bool, 1)

	softTimeoutCtx, softCancel := context.WithTimeout(ctx, time.Duration(softTimeoutMS)*time.Millisecond)
	defer softCancel()

	go func() {
		targetFunc(softTimeoutCtx)
		successChan <- true
	}()

	select {
	case <-successChan:
		// Success
		return true
	case <-time.After(time.Duration(hardTimeoutMS) * time.Millisecond):
		// Timeout
		return false
	}
}

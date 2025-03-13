package util

import (
	"context"
	"testing"
	"time"
)

func TestRunWithTimeoutFullBase(test *testing.T) {
	testCases := []struct {
		softTimeoutMS      int64
		hardTimeoutMS      int64
		waitTimeMS         int64
		observeSoftTimeout bool
		expected           bool
	}{
		// Complete before any timeout.
		{10, 20, 0, true, true},

		// Complete between timeouts, but don't observe soft timeout.
		{1, 1000, 20, false, true},

		// Never finish, but observe the soft timeout.
		{1, 20, 1000, true, true},

		// Never finish, and don't observe the soft timeout.
		{1, 20, 1000, false, false},
	}

	for i, testCase := range testCases {
		testFunc := makeWaitTestFunc(testCase.waitTimeMS, testCase.observeSoftTimeout)
		actual := RunWithTimeoutFull(testCase.softTimeoutMS, testCase.hardTimeoutMS, context.Background(), testFunc)
		if testCase.expected != actual {
			test.Errorf("Case %d: Unexpected result. Expected: '%v', Actual: '%v'.", i, testCase.expected, actual)
			continue
		}
	}
}

func makeWaitTestFunc(timeoutMS int64, observeSoftTimeout bool) func(context.Context) {
	return func(ctx context.Context) {
		softTimeoutChan := make(<-chan struct{})
		if observeSoftTimeout {
			softTimeoutChan = ctx.Done()
		}

		select {
		case <-time.After(time.Duration(timeoutMS) * time.Millisecond):
			return
		case <-softTimeoutChan:
			return
		}
	}
}

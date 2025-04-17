package jobmanager

import (
	"context"
	// "fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/util"
)

var testLockKey string = "test_key"
var testPoolSize int = 1

func TestJobOptionsValidateBase(test *testing.T) {
	testCases := []struct {
		input       *JobOptions
		expected    *JobOptions
		errorString string
	}{
		// Success

		// No context provided
		{
			&JobOptions{
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			&JobOptions{
				Context:  context.Background(),
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				WaitForCompletion: true,
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			&JobOptions{
				WaitForCompletion: true,
				Context:           context.Background(),
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				RetainOriginalContext: true,
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			&JobOptions{
				RetainOriginalContext: true,
				Context:               context.Background(),
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			"",
		},
		// Nil context provided
		{
			&JobOptions{
				Context:  nil,
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			&JobOptions{
				Context:  context.Background(),
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				Context:           nil,
				WaitForCompletion: true,
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			&JobOptions{
				Context:           context.Background(),
				WaitForCompletion: true,
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				Context:               nil,
				RetainOriginalContext: true,
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			&JobOptions{
				Context:               context.Background(),
				RetainOriginalContext: true,
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			"",
		},
		// Context provided
		{
			&JobOptions{
				Context:  context.TODO(),
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			&JobOptions{
				Context:  context.Background(),
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				Context:           context.TODO(),
				WaitForCompletion: true,
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			&JobOptions{
				Context:           context.TODO(),
				WaitForCompletion: true,
				LockKey:           testLockKey,
				PoolSize:          testPoolSize,
			},
			"",
		},
		{
			&JobOptions{
				Context:               context.TODO(),
				RetainOriginalContext: true,
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			&JobOptions{
				Context:               context.TODO(),
				RetainOriginalContext: true,
				LockKey:               testLockKey,
				PoolSize:              testPoolSize,
			},
			"",
		},

		// Errors

		// Nil
		{
			nil,
			nil,
			"Job options are nil.",
		},
		// Bad lock key
		{
			&JobOptions{},
			nil,
			"Cannot have an empty lock key.",
		},
		{
			&JobOptions{
				LockKey: "",
			},
			nil,
			"Cannot have an empty lock key.",
		},
		// Bad pool size
		{
			&JobOptions{
				LockKey: testLockKey,
			},
			nil,
			"Pool size must be positive, got 0.",
		},
		{
			&JobOptions{
				LockKey:  testLockKey,
				PoolSize: 0,
			},
			nil,
			"Pool size must be positive, got 0.",
		},
		{
			&JobOptions{
				LockKey:  testLockKey,
				PoolSize: -1,
			},
			nil,
			"Pool size must be positive, got -1.",
		},
	}

	for i, testCase := range testCases {
		err := testCase.input.Validate()
		if err != nil {
			if testCase.errorString != "" {
				if err.Error() != testCase.errorString {
					test.Errorf("Case %d: Did not get expected error outpout. Expected substring '%s', actual error: '%v'.", i, testCase.errorString, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate '%+v': '%v'.", i, testCase.input, err)
			}

			continue
		}

		if testCase.errorString != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.errorString)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.input) {
			test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.",
				i, testCase.expected, testCase.input)
			continue
		}
	}
}

// TODO: Test caching functions and cache removal functions.
func TestRunJobBase(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	finalExpected := []int{
		1,
		2,
		3,
	}

	workFunc := func(input string) (int, int64, error) {
		return len(input), int64(1), nil
	}

	testCases := []struct {
		input       []string
		output      []int
		workFunc    func(input string) (int, int64, error)
		errorString string
	}{
		// Success
		{input, finalExpected, workFunc, ""},
		{nil, []int{}, workFunc, ""},
		{[]string{}, []int{}, workFunc, ""},

		// Errors
		{input, nil, nil, "Cannot run job with a nil work function."},
		{nil, nil, nil, "Cannot run job with a nil work function."},
	}

	for i, testCase := range testCases {
		options := &JobOptions{
			LockKey:  testLockKey,
			PoolSize: testPoolSize,
		}

		output, numRemaining, runTime, err := RunJob(options, testCase.input, nil, nil, testCase.workFunc)
		if err != nil {
			if testCase.errorString != "" {
				if err.Error() != testCase.errorString {
					test.Errorf("Case %d: Did not get expected error outpout. Expected substring '%s', actual error: '%v'.", i, testCase.errorString, err)
				}
			} else {
				test.Errorf("Case %d: Failed to run initial job: '%v'.", i, err)
			}

			continue
		}

		if testCase.errorString != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.errorString)
			continue
		}

		if len(output) != 0 {
			test.Errorf("Case %d: Unexpected number of initial results. Expected: '0', actual: '%d'.", i, len(output))
			continue
		}

		if numRemaining != len(testCase.input) {
			test.Errorf("Case %d: Unexpected number of initial items remaining. Expected: '%d', actual: '%d'.", i, len(testCase.input), numRemaining)
			continue
		}

		if runTime != 0 {
			test.Errorf("Case %d: Unexpected initial run time. Expected: '0', actual: '%d'.", i, runTime)
			continue
		}

		options.WaitForCompletion = true

		output, numRemaining, runTime, err = RunJob(options, testCase.input, nil, nil, testCase.workFunc)
		if err != nil {
			test.Errorf("Case %d: Failed to run final job: '%v'.", i, err)
			continue
		}

		if numRemaining != 0 {
			test.Errorf("Case %d: Unexpected number of final items remaining. Expected: '0', actual: '%d'.", i, numRemaining)
			continue
		}

		expectedRunTime := int64(len(testCase.input))
		if runTime != expectedRunTime {
			test.Errorf("Case %d: Unexpected final run time. Expected: '%d', actual: '%d'.", i, expectedRunTime, runTime)
			continue
		}

		if !reflect.DeepEqual(testCase.output, output) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(output))
			continue
		}
	}
}

func TestRunJobCancel(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	// Block until the first worker has started.
	workWaitGroup := sync.WaitGroup{}
	workWaitGroup.Add(1)

	workFunc := func(input string) (int, int64, error) {
		// Signal on the first piece of work that we can make sure the workers have started up before we cancel.
		if input == "A" {
			workWaitGroup.Done()
		}

		// Sleep for a really long time (for a test).
		time.Sleep(1 * time.Hour)

		return len(input), int64(1), nil
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	options := &JobOptions{
		Context:           ctx,
		WaitForCompletion: true,
		LockKey:           testLockKey,
		PoolSize:          testPoolSize,
	}

	// Cancel the context as soon as the first worker signals it.
	go func() {
		workWaitGroup.Wait()
		cancelFunc()
	}()

	output, numRemaining, runTime, err := RunJob(options, input, nil, nil, workFunc)
	if err != nil {
		test.Fatalf("Got an unexpected error: '%v'.", err)
	}

	if output != nil {
		test.Fatalf("Got a result when it should have been nil.")
	}

	if numRemaining != 0 {
		test.Fatalf("Got jobs remaining when it should be 0.")
	}

	if runTime != 0 {
		test.Fatalf("Got run time when it should be 0.")
	}
}

package jobmanager

import (
	"context"
	"fmt"
	"reflect"
	"strings"
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
					test.Errorf("Case %d: Did not get expected error output. Expected substring '%s', actual error: '%v'.", i, testCase.errorString, err)
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

	cache := resetCache()

	cacheFunc := func(inputs []string) ([]int, []string, error) {
		outputs := make([]int, 0, len(cache))
		remaining := make([]string, 0, len(inputs))

		for _, input := range inputs {
			output, ok := cache[input]
			if !ok {
				remaining = append(remaining, input)
				continue
			}

			outputs = append(outputs, output)
		}

		return outputs, remaining, nil
	}

	errorCacheFunc := func(inputs []string) ([]int, []string, error) {
		return nil, nil, fmt.Errorf("Crazy cache error!")
	}

	removeCacheFunc := func(inputs []string) error {
		for _, input := range inputs {
			delete(cache, input)
		}

		return nil
	}

	errorRemoveCacheFunc := func(inputs []string) error {
		return fmt.Errorf("Insane cache removal error!")
	}

	workFunc := func(input string) (int, int64, error) {
		return len(input), int64(1), nil
	}

	workFuncWithCache := func(input string) (int, int64, error) {
		cache[input] = len(input)

		return len(input), int64(1), nil
	}

	testCases := []struct {
		input   []string
		options *JobOptions

		initialRemaining int

		initialOutput []int
		finalOutput   []int

		finalRunTime int64

		workFunc         func(input string) (int, int64, error)
		cacheFunc        func(inputs []string) ([]int, []string, error)
		cacheRemovalFunc func(inputs []string) error

		initialErrorSubstring string
		finalErrorSubstring   string

		resetCache      bool
		checkEmptyCache bool
	}{
		// Success

		// Base options
		{
			input:            input,
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
		},
		{
			input:         nil,
			initialOutput: []int{},
			finalOutput:   []int{},
			workFunc:      workFunc,
		},
		{
			input:         []string{},
			initialOutput: []int{},
			finalOutput:   []int{},
			workFunc:      workFunc,
		},

		// Passing a cache function
		{
			input:            input,
			initialOutput:    []int{1, 2},
			initialRemaining: 1,
			finalOutput:      finalExpected,
			finalRunTime:     int64(1),
			workFunc:         workFunc,
			cacheFunc:        cacheFunc,
		},
		{
			input: input,
			options: &JobOptions{
				OverwriteCache: true,
				LockKey:        testLockKey,
				PoolSize:       testPoolSize,
			},
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
			cacheFunc:        cacheFunc,
		},
		{
			input: input,
			options: &JobOptions{
				OverwriteCache: true,
				LockKey:        testLockKey,
				PoolSize:       testPoolSize,
			},
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
			// Won't cause an error because it won't be called.
			cacheFunc: errorCacheFunc,
		},

		// Passing a cache removal function
		{
			input:            input,
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
			cacheRemovalFunc: removeCacheFunc,
		},
		{
			input:            input,
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
			// Won't cause an error because it won't be called.
			cacheRemovalFunc: errorRemoveCacheFunc,
		},

		// Passing cache and cache removal functions
		{
			input:            input,
			initialOutput:    []int{1, 2},
			initialRemaining: 1,
			finalOutput:      finalExpected,
			finalRunTime:     int64(1),
			workFunc:         workFunc,
			cacheFunc:        cacheFunc,
			cacheRemovalFunc: removeCacheFunc,
		},
		{
			input: input,
			options: &JobOptions{
				OverwriteCache: true,
				LockKey:        testLockKey,
				PoolSize:       testPoolSize,
			},
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFunc,
			cacheFunc:        cacheFunc,
			cacheRemovalFunc: removeCacheFunc,
			resetCache:       true,
			checkEmptyCache:  true,
		},
		{
			input:            input,
			initialOutput:    []int{1, 2},
			initialRemaining: 1,
			finalOutput:      finalExpected,
			finalRunTime:     int64(1),
			workFunc:         workFuncWithCache,
			cacheFunc:        cacheFunc,
			cacheRemovalFunc: removeCacheFunc,
			resetCache:       true,
		},
		{
			input: input,
			options: &JobOptions{
				OverwriteCache: true,
				LockKey:        testLockKey,
				PoolSize:       testPoolSize,
			},
			initialOutput:    []int{},
			initialRemaining: len(input),
			finalOutput:      finalExpected,
			finalRunTime:     int64(len(input)),
			workFunc:         workFuncWithCache,
			cacheFunc:        cacheFunc,
			cacheRemovalFunc: removeCacheFunc,
			resetCache:       true,
			checkEmptyCache:  true,
		},

		// Errors

		// Nil work function
		{
			input:                 input,
			initialOutput:         []int{},
			workFunc:              nil,
			initialErrorSubstring: "Cannot run job with a nil work function.",
		},
		{
			input:                 nil,
			initialOutput:         []int{},
			workFunc:              nil,
			initialErrorSubstring: "Cannot run job with a nil work function.",
		},

		// Bad cache function
		{
			input:                 input,
			initialOutput:         []int{},
			workFunc:              workFunc,
			cacheFunc:             errorCacheFunc,
			initialErrorSubstring: "Crazy cache error!",
		},

		// Bad cache removal function
		{
			input: input,
			options: &JobOptions{
				OverwriteCache: true,
				LockKey:        testLockKey,
				PoolSize:       testPoolSize,
			},
			initialOutput:       []int{},
			initialRemaining:    len(input),
			workFunc:            workFunc,
			cacheRemovalFunc:    errorRemoveCacheFunc,
			finalErrorSubstring: "Insane cache removal error!",
		},
	}

	for i, testCase := range testCases {
		if testCase.options == nil {
			testCase.options = &JobOptions{
				LockKey:  testLockKey,
				PoolSize: testPoolSize,
			}
		}

		testCase.options.WaitForCompletion = false

		output, numRemaining, runTime, err := RunJob(testCase.options, testCase.input, testCase.cacheFunc, testCase.cacheRemovalFunc, testCase.workFunc)
		if err != nil {
			if testCase.initialErrorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.initialErrorSubstring) {
					test.Errorf("Case %d: Did not get expected error output on initial run. Expected substring: '%s', actual error: '%v'.", i, testCase.initialErrorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to run initial job: '%v'.", i, err)
			}

			continue
		}

		if testCase.initialErrorSubstring != "" {
			test.Errorf("Case %d: Did not get expected initial error: '%s'.", i, testCase.initialErrorSubstring)
			continue
		}

		if numRemaining != testCase.initialRemaining {
			test.Errorf("Case %d: Unexpected number of initial items remaining. Expected: '%d', actual: '%d'.", i, testCase.initialRemaining, numRemaining)
			continue
		}

		if runTime != 0 {
			test.Errorf("Case %d: Unexpected initial run time. Expected: '0', actual: '%d'.", i, runTime)
			continue
		}

		if !reflect.DeepEqual(output, testCase.initialOutput) {
			test.Errorf("Case %d: Unexpected initial results. Expected: '%v', actual: '%v'.", i, testCase.initialOutput, output)
			continue
		}

		testCase.options.WaitForCompletion = true

		output, numRemaining, runTime, err = RunJob(testCase.options, testCase.input, testCase.cacheFunc, testCase.cacheRemovalFunc, testCase.workFunc)
		if err != nil {
			if testCase.finalErrorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.finalErrorSubstring) {
					test.Errorf("Case %d: Did not get expected error output on final run. Expected substring: '%s', actual error: '%v'.", i, testCase.finalErrorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to run final job: '%v'.", i, err)
			}

			continue
		}

		if testCase.finalErrorSubstring != "" {
			test.Errorf("Case %d: Did not get expected final error: '%s'.", i, testCase.finalErrorSubstring)
			continue
		}

		if numRemaining != 0 {
			test.Errorf("Case %d: Unexpected number of final items remaining. Expected: '0', actual: '%d'.", i, numRemaining)
			continue
		}

		if runTime != testCase.finalRunTime {
			test.Errorf("Case %d: Unexpected final run time. Expected: '%d', actual: '%d'.", i, testCase.finalRunTime, runTime)
			continue
		}

		if !reflect.DeepEqual(testCase.finalOutput, output) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalOutput), util.MustToJSONIndent(output))
			continue
		}

		if testCase.checkEmptyCache && !testCase.resetCache {
			test.Errorf("Case %d: A test that checks for an empty cache must reset the cache.", i)
			cache = resetCache()
			continue
		}

		if testCase.resetCache {
			if testCase.checkEmptyCache {
				testCase.options.WaitForCompletion = false

				output, numRemaining, runTime, err = RunJob(testCase.options, testCase.input, testCase.cacheFunc, testCase.cacheRemovalFunc, testCase.workFunc)
				if err != nil {
					test.Errorf("Case %d: Failed to check for an empty cache: '%v'.", i, err)
					cache = resetCache()
					continue
				}

				if runTime != 0 {
					test.Errorf("Case %d: Unexpected run time during cache check. Expected: '0', actual: '%d'.", i, runTime)
					cache = resetCache()
					continue
				}

				if numRemaining != len(testCase.input) {
					test.Errorf("Case %d: Unexpected number of items remaining during cache check. Expected: '%d', actual: '%d'.",
						i, len(testCase.input), numRemaining)
					cache = resetCache()
					continue
				}

				if !reflect.DeepEqual(output, []int{}) {
					test.Errorf("Case %d: Unexpected output during cache check. Expected: '%v', actual: '%v'.", i, []int{}, output)
					cache = resetCache()
					continue
				}
			}

			cache = resetCache()
		}
	}
}

func TestRunJobCancel(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	// Block until the initial worker has started.
	workWaitGroup := sync.WaitGroup{}
	workWaitGroup.Add(1)

	workFunc := func(input string) (int, int64, error) {
		// Signal on the initial piece of work that we can make sure the workers have started up before we cancel.
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

	// Cancel the context as soon as the initial worker signals it.
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

func resetCache() map[string]int {
	return map[string]int{
		"A":  1,
		"BB": 2,
	}
}

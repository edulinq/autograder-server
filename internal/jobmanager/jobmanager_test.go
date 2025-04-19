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

var workFunc = func(input string) (int, int64, error) {
	return len(input), int64(1), nil
}

func TestJobValidateBase(test *testing.T) {
	testCases := []struct {
		input       *Job[string, int]
		expected    *Job[string, int]
		errorString string
	}{
		// Success

		// No Context Provided
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:  context.Background(),
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:               context.Background(),
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			"",
		},
		// Nil context provided
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:  nil,
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:  context.Background(),
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:           nil,
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:               nil,
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:               context.Background(),
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			"",
		},
		// Context provided
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:  context.TODO(),
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					// Swap the context to background when not waiting for completion.
					Context:  context.Background(),
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
					PoolSize:          testPoolSize,
					LockKey:           testLockKey,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:               context.TODO(),
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					Context:               context.TODO(),
					RetainOriginalContext: true,
					PoolSize:              testPoolSize,
					LockKey:               testLockKey,
				},
			},
			"",
		},

		// Errors

		// Nil
		{
			nil,
			nil,
			"Job is nil.",
		},
		// Bad Work Function
		{
			&Job[string, int]{},
			nil,
			"Job cannot have a nil work function.",
		},
		{
			&Job[string, int]{
				WorkFunc: nil,
			},
			nil,
			"Job cannot have a nil work function.",
		},
		// Bad Pool Size
		{
			&Job[string, int]{
				WorkFunc: workFunc,
			},
			nil,
			"Pool size must be positive, got 0.",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					PoolSize: 0,
				},
			},
			nil,
			"Pool size must be positive, got 0.",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				JobOptions: JobOptions{
					PoolSize: -1,
				},
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

		if testCase.input.WorkFunc == nil {
			test.Errorf("Case %d: Found non nil work func.", i)
			continue
		}

		// Clear work functions for comparison.
		testCase.input.WorkFunc = nil
		testCase.expected.WorkFunc = nil

		if !reflect.DeepEqual(testCase.expected, testCase.input) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(testCase.input))
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

	storage := resetStorage()

	retrieveFunc := func(inputs []string) ([]int, []string, error) {
		outputs := make([]int, 0, len(storage))
		remaining := make([]string, 0, len(inputs))

		for _, input := range inputs {
			output, ok := storage[input]
			if !ok {
				remaining = append(remaining, input)
				continue
			}

			outputs = append(outputs, output)
		}

		return outputs, remaining, nil
	}

	errorRetrieveFunc := func(_ []string) ([]int, []string, error) {
		return nil, nil, fmt.Errorf("Crazy retrieval error!")
	}

	removeStorageFunc := func(inputs []string) error {
		for _, input := range inputs {
			delete(storage, input)
		}

		return nil
	}

	errorRemoveStorageFunc := func(_ []string) error {
		return fmt.Errorf("Insane storage removal error!")
	}

	workFuncWithStorage := func(input string) (int, int64, error) {
		storage[input] = len(input)

		return len(input), int64(1), nil
	}

	testCases := []struct {
		job Job[string, int]

		initialOutput JobOutput[string, int]
		finalOutput   JobOutput[string, int]

		initialErrorSubstring string
		finalErrorSubstring   string

		resetStorage      bool
		checkEmptyStorage bool
	}{
		// Success

		// Base Options
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkFunc:  workFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
		},
		{
			job: Job[string, int]{
				WorkItems: nil,
				WorkFunc:  workFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				RunTime:        int64(0),
			},
		},
		{
			job: Job[string, int]{
				WorkItems: []string{},
				WorkFunc:  workFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				RunTime:        int64(0),
			},
		},

		// Passing A Retrieval Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				RetrieveFunc: retrieveFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(1),
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				RetrieveFunc: retrieveFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
					PoolSize:         testPoolSize,
					LockKey:          testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkFunc:  workFunc,
				// Won't cause an error because it won't be called.
				RetrieveFunc: errorRetrieveFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
					PoolSize:         testPoolSize,
					LockKey:          testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
		},

		// Passing A Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkFunc:  workFunc,
				// Won't cause an error because it won't be called.
				RemoveStorageFunc: errorRemoveStorageFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
		},

		// Passing Retrieval And Storage Removal Functions
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(1),
			},
		},
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
					PoolSize:         testPoolSize,
					LockKey:          testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
			resetStorage:      true,
			checkEmptyStorage: true,
		},
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFuncWithStorage,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
					LockKey:  testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(1),
			},
			resetStorage: true,
		},
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFuncWithStorage,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
					PoolSize:         testPoolSize,
					LockKey:          testLockKey,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalOutput: JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
			},
			resetStorage:      true,
			checkEmptyStorage: true,
		},

		// Errors

		// Nil Work Function
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkFunc:  nil,
			},
			initialErrorSubstring: "Job cannot have a nil work function.",
		},
		{
			job: Job[string, int]{
				WorkItems: nil,
				WorkFunc:  nil,
			},
			initialErrorSubstring: "Job cannot have a nil work function.",
		},

		// Bad Storage Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				RetrieveFunc: errorRetrieveFunc,
				JobOptions: JobOptions{
					PoolSize: testPoolSize,
				},
			},
			initialErrorSubstring: "Crazy retrieval error!",
		},

		// Bad Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				RemoveStorageFunc: errorRemoveStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
					PoolSize:         testPoolSize,
				},
			},
			initialOutput: JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
			},
			finalErrorSubstring: "Insane storage removal error!",
		},
	}

	for i, testCase := range testCases {
		testCase.job.WaitForCompletion = false

		output, err := testCase.job.Run()
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

		// Set the done channel to pass the equality check.
		testCase.initialOutput.Done = output.Done

		if !reflect.DeepEqual(output, testCase.initialOutput) {
			test.Errorf("Case %d: Unexpected initial results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.initialOutput), util.MustToJSONIndent(output))
			continue
		}

		testCase.job.WaitForCompletion = true

		output, err = testCase.job.Run()
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

		// Set the done channel to pass the equality check.
		testCase.finalOutput.Done = output.Done

		if !reflect.DeepEqual(output, testCase.finalOutput) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalOutput), util.MustToJSONIndent(output))
			continue
		}

		if testCase.checkEmptyStorage && !testCase.resetStorage {
			test.Errorf("Case %d: A test that checks for an empty storage must reset the storage.", i)
			storage = resetStorage()
			continue
		}

		if testCase.resetStorage {
			if testCase.checkEmptyStorage {
				testCase.job.WaitForCompletion = false

				output, err = testCase.job.Run()
				if err != nil {
					test.Errorf("Case %d: Failed to check for an empty storage: '%v'.", i, err)
					storage = resetStorage()
					continue
				}

				expected := JobOutput[string, int]{
					ResultItems:    []int{},
					RemainingItems: input,
					RunTime:        int64(0),
					// Set the done channel to pass the equality check.
					Done: output.Done,
				}

				if !reflect.DeepEqual(output, expected) {
					test.Errorf("Case %d: Unexpected output during storage check. Expected: '%v', actual: '%v'.", i, expected, output)
					storage = resetStorage()
					continue
				}
			}

			storage = resetStorage()
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

	sleepWorkFunc := func(input string) (int, int64, error) {
		// Signal on the initial piece of work that we can make sure the workers have started up before we cancel.
		if input == "A" {
			workWaitGroup.Done()
		}

		// Sleep for a really long time (for a test).
		time.Sleep(1 * time.Hour)

		return len(input), int64(1), nil
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	job := &Job[string, int]{
		WorkItems: input,
		WorkFunc:  sleepWorkFunc,
		JobOptions: JobOptions{
			Context:           ctx,
			WaitForCompletion: true,
			PoolSize:          testPoolSize,
			LockKey:           testLockKey,
		},
	}

	// Cancel the context as soon as the initial worker signals it.
	go func() {
		workWaitGroup.Wait()
		cancelFunc()
	}()

	output, err := job.Run()
	if err != nil {
		test.Fatalf("Got an unexpected error: '%v'.", err)
	}

	if !reflect.DeepEqual(output, JobOutput[string, int]{}) {
		test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(JobOutput[string, int]{}), util.MustToJSONIndent(output))
	}
}

func TestRunJobChannel(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	job := &Job[string, int]{
		WorkItems: input,
		WorkFunc:  workFunc,
		JobOptions: JobOptions{
			WaitForCompletion: false,
			PoolSize:          testPoolSize,
			LockKey:           testLockKey,
		},
	}

	output, err := job.Run()
	if err != nil {
		test.Fatalf("Failed to run job: '%v'.", err)
	}

	// Wait for the worker to signal the job is done.
	<-output.Done

	expected := JobOutput[string, int]{
		ResultItems:    []int{1, 2, 3},
		RemainingItems: []string{},
		RunTime:        int64(len(input)),
		Done:           job.Done,
	}

	// Must check the job object itself for updates.
	// The output variable is returned before the work is done.
	if !reflect.DeepEqual(job.JobOutput, expected) {
		test.Fatalf("Unexpected output. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(job.JobOutput))
	}
}

func resetStorage() map[string]int {
	return map[string]int{
		"A":  1,
		"BB": 2,
	}
}

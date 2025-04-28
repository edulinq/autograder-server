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

var workFunc = func(input string) (int, error) {
	return len(input), nil
}

type printableJob[InputType any, OutputType any] struct {
	*JobOptions

	Context context.Context `json:"context"`

	PoolSize int `json:"pool-size"`

	LockKey string `json:"lock-key"`

	WorkItems []InputType `json:"work-items"`
}

func (this *Job[InputType, OutputType]) toPrintableJob() *printableJob[InputType, OutputType] {
	if this == nil {
		return nil
	}

	return &printableJob[InputType, OutputType]{
		JobOptions: this.JobOptions,
		Context:    this.Context,
		PoolSize:   this.PoolSize,
		LockKey:    this.LockKey,
		WorkItems:  this.WorkItems,
	}
}

type printableJobOutput[InputType any, OutputType any] struct {
	Error error `json:"error"`

	WorkErrors map[int]error `json:"work-errors"`

	ResultItems []OutputType `json:"result-items"`

	RemainingItems []InputType `json:"remaining-items"`

	RunTime int64 `json:"run-time"`
}

func (this *JobOutput[InputType, OutputType]) toPrintableJobOutput() *printableJobOutput[InputType, OutputType] {
	if this == nil {
		return nil
	}

	return &printableJobOutput[InputType, OutputType]{
		Error:          this.Error,
		WorkErrors:     this.WorkErrors,
		ResultItems:    this.ResultItems,
		RemainingItems: this.RemainingItems,
		RunTime:        this.RunTime,
	}
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
				WorkFunc:   workFunc,
				PoolSize:   testPoolSize,
				LockKey:    testLockKey,
				JobOptions: &JobOptions{},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context: context.Background(),
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
				},
			},
			"",
		},

		// Nil context provided
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context: nil,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context: context.Background(),
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context:           nil,
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
				},
			},
			"",
		},

		// Context provided
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context: context.TODO(),
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context: context.TODO(),
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: &JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
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
				PoolSize: 0,
			},
			nil,
			"Pool size must be positive, got 0.",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: -1,
			},
			nil,
			"Pool size must be positive, got -1.",
		},
		// Nil Job Options
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
			},
			nil,
			"Job options are nil.",
		},
		{
			&Job[string, int]{
				WorkFunc:   workFunc,
				PoolSize:   testPoolSize,
				JobOptions: nil,
			},
			nil,
			"Job options are nil.",
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
			test.Errorf("Case %d: Found a nil work func after validation.", i)
			continue
		}

		// Clear work functions for comparison.
		testCase.input.WorkFunc = nil
		testCase.expected.WorkFunc = nil

		if !reflect.DeepEqual(testCase.expected, testCase.input) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected.toPrintableJob()), util.MustToJSONIndent(testCase.input.toPrintableJob()))
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

	errorRemoveFunc := func(_ []string) error {
		return fmt.Errorf("Insane storage removal error!")
	}

	storageFunc := func(results []int) error {
		for _, result := range results {
			switch result {
			case 1:
				storage["A"] = 1
			case 2:
				storage["BB"] = 2
			case 3:
				storage["CCC"] = 3
			default:
				storage["unknown"] = result
			}
		}

		return nil
	}

	testCases := []struct {
		job Job[string, int]

		initialOutput *JobOutput[string, int]
		finalOutput   *JobOutput[string, int]

		errorSubstring string

		expectedStorage map[string]int
	}{
		// Success

		// Base Options
		{
			job: Job[string, int]{
				WorkItems:  input,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  nil,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  []string{},
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},

		// Passing A Retrieval Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: retrieveFunc,
				JobOptions:   &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: retrieveFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},

		// Passing A Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:  input,
				RemoveFunc: removeStorageFunc,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				// Won't be called due to OverwriteRecords.
				RemoveFunc: errorRemoveFunc,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},

		// Passing Retrieval And Storage Removal Functions
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: retrieveFunc,
				// Storage removal is not called.
				RemoveFunc: removeStorageFunc,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: retrieveFunc,
				RemoveFunc:   removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			expectedStorage: map[string]int{},
		},

		// Storage Functions
		{
			job: Job[string, int]{
				WorkItems:  input,
				StoreFunc:  storageFunc,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			expectedStorage: map[string]int{
				"A":   1,
				"BB":  2,
				"CCC": 3,
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				StoreFunc: storageFunc,
				JobOptions: &JobOptions{
					DryRun: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			expectedStorage: map[string]int{
				"A":  1,
				"BB": 2,
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				StoreFunc:    storageFunc,
				RetrieveFunc: retrieveFunc,
				RemoveFunc:   removeStorageFunc,
				JobOptions:   &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			expectedStorage: map[string]int{
				"A":   1,
				"BB":  2,
				"CCC": 3,
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				StoreFunc:    storageFunc,
				RetrieveFunc: retrieveFunc,
				RemoveFunc:   removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			// Storage happens after record removal.
			expectedStorage: map[string]int{
				"A":   1,
				"BB":  2,
				"CCC": 3,
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				StoreFunc:    storageFunc,
				RetrieveFunc: retrieveFunc,
				RemoveFunc:   removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
					DryRun:           true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				WorkErrors:     map[int]error{},
			},
			expectedStorage: map[string]int{
				"A":  1,
				"BB": 2,
			},
		},

		// Errors

		// Bad Storage Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: errorRetrieveFunc,
				JobOptions:   &JobOptions{},
			},
			errorSubstring: "Crazy retrieval error!",
		},

		// Bad Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:  input,
				RemoveFunc: errorRemoveFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
				},
			},
			errorSubstring: "Insane storage removal error!",
		},
	}

	for i, testCase := range testCases {
		testCase.job.WorkFunc = workFunc
		testCase.job.PoolSize = testPoolSize
		testCase.job.LockKey = testLockKey

		testCase.job.WaitForCompletion = false
		testCase.job.ReturnIncompleteResults = true

		output := testCase.job.Run()

		// Clear channel and run time for comparison.
		output.Done = nil
		output.RunTime = 0

		if output.Error != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(output.Error.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output on initial run. Expected substring: '%s', actual error: '%v'.", i, testCase.errorSubstring, output.Error)
				}
			} else {
				test.Errorf("Case %d: Failed to run initial job: '%v'.", i, output.Error)
			}

			storage = resetStorage()
			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected initial error: '%s'.", i, testCase.errorSubstring)
			storage = resetStorage()
			continue
		}

		if !reflect.DeepEqual(output, testCase.initialOutput) {
			test.Errorf("Case %d: Unexpected initial results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.initialOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
			storage = resetStorage()
			continue
		}

		testCase.job.WaitForCompletion = true

		output = testCase.job.Run()
		if output.Error != nil {
			test.Errorf("Case %d: Failed to run final job: '%v'.", i, output.Error)
			storage = resetStorage()
			continue
		}

		// Clear channel and run time for comparison.
		output.Done = nil
		output.RunTime = 0

		if !reflect.DeepEqual(output, testCase.finalOutput) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
			storage = resetStorage()
			continue
		}

		if testCase.expectedStorage == nil {
			testCase.expectedStorage = map[string]int{
				"A":  1,
				"BB": 2,
			}
		}

		// Sleep to let the writes to the storage complete.
		time.Sleep(time.Duration(2) * time.Millisecond)

		if !reflect.DeepEqual(storage, testCase.expectedStorage) {
			test.Errorf("Case %d: Unexpected storage results after final run. Expected: '%v', actual: '%v'.", i, testCase.expectedStorage, storage)
			storage = resetStorage()
			continue
		}

		storage = resetStorage()
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

	sleepWorkFunc := func(input string) (int, error) {
		// Signal on the initial piece of work that we can make sure the workers have started up before we cancel.
		if input == "A" {
			workWaitGroup.Done()
		}

		// Sleep for a really long time (for a test).
		time.Sleep(1 * time.Hour)

		return len(input), nil
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	job := &Job[string, int]{
		WorkItems: input,
		WorkFunc:  sleepWorkFunc,
		PoolSize:  testPoolSize,
		LockKey:   testLockKey,
		JobOptions: &JobOptions{
			Context:           ctx,
			WaitForCompletion: true,
		},
	}

	// Cancel the context as soon as the initial worker signals it.
	go func() {
		workWaitGroup.Wait()
		cancelFunc()
	}()

	output := job.Run()
	if output != nil {
		test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(nil), util.MustToJSONIndent(output.toPrintableJobOutput()))
	}
}

func TestRunJobChannel(test *testing.T) {
	input := []string{
		"A",
		"BB",
		"CCC",
	}

	job := &Job[string, int]{
		WorkItems:               input,
		WorkFunc:                workFunc,
		PoolSize:                testPoolSize,
		LockKey:                 testLockKey,
		ReturnIncompleteResults: false,
		JobOptions: &JobOptions{
			WaitForCompletion: false,
		},
	}

	output := job.Run()
	if output.Error != nil {
		test.Fatalf("Failed to run job: '%v'.", output.Error)
	}

	// Wait for the worker to signal the job is done.
	<-output.Done

	expected := &JobOutput[string, int]{
		ResultItems:    []int{1, 2, 3},
		RemainingItems: []string{},
		RunTime:        output.RunTime,
		WorkErrors:     map[int]error{},
		Done:           output.Done,
	}

	if !reflect.DeepEqual(output, expected) {
		test.Fatalf("Unexpected output. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
	}
}

func resetStorage() map[string]int {
	return map[string]int{
		"A":  1,
		"BB": 2,
	}
}

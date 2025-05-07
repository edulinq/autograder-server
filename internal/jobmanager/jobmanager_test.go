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

var input []string = []string{
	"A",
	"BB",
	"CCC",
}

var workFunc = func(input string) (int, error) {
	return len(input), nil
}

type printableJob[InputType any, OutputType any] struct {
	*JobOptions

	// Overwrite the JSON tag to display the context.
	Context context.Context `json:"context"`

	ReturnIncompleteResults bool `json:"return-incomplete-results"`

	PoolSize int `json:"pool-size"`

	LockKey string `json:"lock-key"`

	WorkItems []InputType `json:"work-items"`
}

func (this *Job[InputType, OutputType]) toPrintableJob() *printableJob[InputType, OutputType] {
	if this == nil {
		return nil
	}

	return &printableJob[InputType, OutputType]{
		JobOptions:              this.JobOptions,
		Context:                 this.Context,
		ReturnIncompleteResults: this.ReturnIncompleteResults,
		PoolSize:                this.PoolSize,
		LockKey:                 this.LockKey,
		WorkItems:               this.WorkItems,
	}
}

type printableJobOutput[InputType comparable, OutputType any] struct {
	Canceled bool `json:"canceled"`

	Error error `json:"error"`

	WorkErrors map[InputType]error `json:"work-errors"`

	ResultItems map[InputType]OutputType `json:"result-items"`

	RemainingItems []InputType `json:"remaining-items"`

	RunTime int64 `json:"run-time"`
}

func (this *JobOutput[InputType, OutputType]) toPrintableJobOutput() *printableJobOutput[InputType, OutputType] {
	if this == nil {
		return nil
	}

	return &printableJobOutput[InputType, OutputType]{
		Canceled:       this.Canceled,
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

		// Nil Context Provided
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
					test.Errorf("Case %d: Did not get expected error output. Expected substring: '%s', actual error: '%s'.", i, testCase.errorString, err.Error())
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
	finalExpected := map[string]int{
		"A":   1,
		"BB":  2,
		"CCC": 3,
	}

	// A basic cache to test caching functionality.
	storage := map[string]int{}

	retrieveFunc := func(_ []string) (map[string]int, error) {
		results := make(map[string]int, len(storage))

		for input, output := range storage {
			results[input] = output
		}

		return results, nil
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
				WorkItems:  nil,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: nil,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: nil,
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  []string{},
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: []string{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: []string{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  input,
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems: map[string]int{
					"A":  1,
					"BB": 2,
				},
				RemainingItems: []string{"CCC"},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
		},

		// Passing A Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems: input,
				// Storage removal is not called when not overwriting records.
				RemoveFunc: removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: false,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  input,
				RemoveFunc: removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
			expectedStorage: map[string]int{},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				// Storage removal is not called when not overwriting records.
				RemoveFunc: errorRemoveFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: false,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
		},

		// Passing Retrieval And Storage Removal Functions
		{
			job: Job[string, int]{
				WorkItems:    input,
				RetrieveFunc: retrieveFunc,
				// Storage removal is not called when not overwriting records.
				RemoveFunc: removeStorageFunc,
				JobOptions: &JobOptions{
					OverwriteRecords: false,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems: map[string]int{
					"A":  1,
					"BB": 2,
				},
				RemainingItems: []string{"CCC"},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
			expectedStorage: map[string]int{
				"A":  1,
				"BB": 2,
			},
		},

		// Retrieval, Removal, And Storage Functions
		{
			job: Job[string, int]{
				WorkItems:    input,
				StoreFunc:    storageFunc,
				RetrieveFunc: retrieveFunc,
				JobOptions:   &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems: map[string]int{
					"A":  1,
					"BB": 2,
				},
				RemainingItems: []string{"CCC"},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
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
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
			expectedStorage: map[string]int{
				"A":  1,
				"BB": 2,
			},
		},

		// Work Item Key Function
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkItemKeyFunc: func(input string) string {
					return input
				},
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkItemKeyFunc: func(_ string) string {
					return "serial_test_key"
				},
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
			},
		},

		// On Success Function
		{
			job: Job[string, int]{
				WorkItems: input,
				OnSuccess: func(output JobOutput[string, int]) {
					for i, result := range output.ResultItems {
						output.ResultItems[i] = result * 2
					}
				},
				JobOptions: &JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    map[string]int{},
				RemainingItems: input,
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems: map[string]int{
					"A":   2,
					"BB":  4,
					"CCC": 6,
				},
				RemainingItems: []string{},
			},
		},

		// Errors

		// Bad Storage Function
		{
			job: Job[string, int]{
				WorkItems: input,
				RetrieveFunc: func(_ []string) (map[string]int, error) {
					return nil, fmt.Errorf("Crazy retrieval error!")
				},
				JobOptions: &JobOptions{},
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
		// Set defualt job config.
		testCase.job.WorkFunc = workFunc
		testCase.job.PoolSize = testPoolSize
		testCase.job.LockKey = testLockKey

		testCase.job.WaitForCompletion = false
		testCase.job.ReturnIncompleteResults = true

		storage = resetStorage()

		output := testCase.job.Run()
		if output.Error != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(output.Error.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output on initial run. Expected substring: '%s', actual error: '%v'.", i, testCase.errorSubstring, output.Error)
				}
			} else {
				test.Errorf("Case %d: Failed to run initial job: '%s'.", i, output.Error.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected initial error: '%s'.", i, testCase.errorSubstring)
			continue
		}

		// Clear channel and run time for comparison.
		output.Done = nil
		output.RunTime = 0

		// Set default error output for successful test cases.
		testCase.initialOutput.WorkErrors = map[string]error{}

		if !reflect.DeepEqual(output, testCase.initialOutput) {
			test.Errorf("Case %d: Unexpected initial results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.initialOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
			continue
		}

		testCase.job.WaitForCompletion = true

		output = testCase.job.Run()
		if output.Error != nil {
			test.Errorf("Case %d: Failed to run final job: '%s'.", i, output.Error.Error())
			continue
		}

		// Clear channel and run time for comparison.
		output.Done = nil
		output.RunTime = 0

		// Set default error output for successful test cases.
		testCase.finalOutput.WorkErrors = map[string]error{}

		if !reflect.DeepEqual(output, testCase.finalOutput) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
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
			test.Errorf("Case %d: Unexpected storage results. Expected: '%v', actual: '%v'.", i, testCase.expectedStorage, storage)
			continue
		}
	}
}

func TestRunJobCancel(test *testing.T) {
	// Block until the initial worker has started.
	workWaitGroup := sync.WaitGroup{}
	workWaitGroup.Add(1)

	sleepWorkFunc := func(input string) (int, error) {
		// Allow the first input to return a result to test cleanup.
		if input == "A" {
			return len(input), nil
		}

		// Signal on the second piece of work so that we can make sure the workers have started up before we cancel.
		if input == "BB" {
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

	// Cancel the context as soon as the worker signals it.
	go func() {
		workWaitGroup.Wait()
		cancelFunc()
	}()

	expectedOutput := &JobOutput[string, int]{
		Canceled:       true,
		Error:          fmt.Errorf("Job was canceled: 'context canceled'."),
		WorkErrors:     map[string]error{},
		ResultItems:    map[string]int{},
		RemainingItems: []string{},
	}

	output := job.Run()
	if output.Error.Error() != expectedOutput.Error.Error() {
		test.Fatalf("Unexpected error. Expected: '%s', actual: '%s'.",
			expectedOutput.Error.Error(), output.Error.Error())
	}

	// Clear done channel and errors for comparison check.
	output.Done = nil
	output.Error = nil
	expectedOutput.Error = nil

	if !reflect.DeepEqual(output, expectedOutput) {
		test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
	}
}

func TestRunJobChannel(test *testing.T) {
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
		ResultItems: map[string]int{
			"A":   1,
			"BB":  2,
			"CCC": 3,
		},
		RemainingItems: []string{},
		RunTime:        output.RunTime,
		WorkErrors:     map[string]error{},
		Done:           output.Done,
	}

	if !reflect.DeepEqual(output, expected) {
		test.Fatalf("Unexpected output. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
	}
}

func TestBadWorkFunc(test *testing.T) {
	job := &Job[string, int]{
		WorkItems: input,
		WorkFunc: func(input string) (int, error) {
			return 0, fmt.Errorf("Sneaky work error.")
		},
		PoolSize:                testPoolSize,
		LockKey:                 testLockKey,
		ReturnIncompleteResults: true,
		JobOptions: &JobOptions{
			WaitForCompletion: false,
		},
	}

	// The incomplete results should not have any errors.
	output := job.Run()
	if output.Error != nil {
		test.Fatalf("Failed to run job: '%s'.", output.Error.Error())
	}

	job.WaitForCompletion = true

	expectedOutput := &JobOutput[string, int]{
		Error:          fmt.Errorf("Failed to complete work for '%d' items.", len(input)),
		RemainingItems: []string{},
		ResultItems:    map[string]int{},
		WorkErrors: map[string]error{
			"A":   fmt.Errorf("Failed to perform individual work on item 'A': 'Sneaky work error.'."),
			"BB":  fmt.Errorf("Failed to perform individual work on item 'BB': 'Sneaky work error.'."),
			"CCC": fmt.Errorf("Failed to perform individual work on item 'CCC': 'Sneaky work error.'."),
		},
	}

	output = job.Run()
	if output.Error == nil {
		test.Fatalf("Did not get expected error. Expected: '%s'.", expectedOutput.Error.Error())
	}

	if output.Error.Error() != expectedOutput.Error.Error() {
		test.Fatalf("Unexpected error. Expected: '%s', actual: '%s'.",
			expectedOutput.Error.Error(), output.Error.Error())
	}

	if len(output.WorkErrors) != len(expectedOutput.WorkErrors) {
		test.Fatalf("Unexpected number of work errors. Expected: '%v', actual: '%v'.",
			expectedOutput.WorkErrors, output.WorkErrors)
	}

	for item, expectedError := range expectedOutput.WorkErrors {
		actualError, ok := output.WorkErrors[item]
		if !ok {
			test.Fatalf("Unable to find expected error for item '%v': '%s'.", item, expectedError.Error())
		}

		if expectedError.Error() != actualError.Error() {
			test.Fatalf("Unexpected work error for item '%v'. Expected: '%s', actual: '%s'.",
				item, expectedError.Error(), actualError.Error())
		}
	}

	// Clear done channel and errors for comparison check.
	output.Done = nil
	expectedOutput.Done = nil

	output.Error = nil
	expectedOutput.Error = nil

	output.WorkErrors = nil
	expectedOutput.WorkErrors = nil

	if !reflect.DeepEqual(output, expectedOutput) {
		test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
	}
}

func resetStorage() map[string]int {
	return map[string]int{
		"A":  1,
		"BB": 2,
	}
}

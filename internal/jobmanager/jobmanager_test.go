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
	JobOptions

	Context context.Context `json:"context"`

	printableJobOutput[InputType, OutputType]

	PoolSize int `json:"pool-size"`

	LockKey string `json:"lock-key"`

	WorkItems []InputType `json:"work-items"`
}

func (this *Job[InputType, OutputType]) toPrintableJob() *printableJob[InputType, OutputType] {
	if this == nil {
		return nil
	}

	return &printableJob[InputType, OutputType]{
		JobOptions:         this.JobOptions,
		Context:            this.Context,
		printableJobOutput: *this.JobOutput.toPrintableJobOutput(),
		PoolSize:           this.PoolSize,
		LockKey:            this.LockKey,
		WorkItems:          this.WorkItems,
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
				JobOptions: JobOptions{},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
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
				JobOptions: JobOptions{
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					RetainOriginalContext: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:               context.Background(),
					RetainOriginalContext: true,
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
				JobOptions: JobOptions{
					Context: nil,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
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
				JobOptions: JobOptions{
					Context:           nil,
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:           context.Background(),
					WaitForCompletion: true,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:               nil,
					RetainOriginalContext: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:               context.Background(),
					RetainOriginalContext: true,
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
				JobOptions: JobOptions{
					Context: context.TODO(),
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					// Swap the context to background when not waiting for completion.
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
				JobOptions: JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:           context.TODO(),
					WaitForCompletion: true,
				},
			},
			"",
		},
		{
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:               context.TODO(),
					RetainOriginalContext: true,
				},
			},
			&Job[string, int]{
				WorkFunc: workFunc,
				PoolSize: testPoolSize,
				LockKey:  testLockKey,
				JobOptions: JobOptions{
					Context:               context.TODO(),
					RetainOriginalContext: true,
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
			if &testCase.expected.WorkFunc != &testCase.input.WorkFunc {
				test.Errorf("Case %d: Wrong work func. Expected address: '%v', actual address: '%v'.", i, &testCase.expected.WorkFunc, &testCase.input.WorkFunc)
			}

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

	retrieveFunc := func(inputs []string) ([]int, []string, []string, error) {
		outputs := make([]int, 0, len(storage))
		complete := make([]string, 0, len(storage))
		remaining := make([]string, 0, len(inputs))

		for _, input := range inputs {
			output, ok := storage[input]
			if !ok {
				remaining = append(remaining, input)
				continue
			}

			outputs = append(outputs, output)
			complete = append(complete, input)
		}

		return outputs, complete, remaining, nil
	}

	errorRetrieveFunc := func(_ []string) ([]int, []string, []string, error) {
		return nil, nil, nil, fmt.Errorf("Crazy retrieval error!")
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

	workFuncWithStorage := func(input string) (int, error) {
		storage[input] = len(input)

		return len(input), nil
	}

	testCases := []struct {
		job Job[string, int]

		initialOutput *JobOutput[string, int]
		finalOutput   *JobOutput[string, int]

		errorSubstring string

		resetStorage      bool
		checkEmptyStorage bool
	}{
		// Success

		// Base Options
		{
			job: Job[string, int]{
				WorkItems:  input,
				WorkFunc:   workFunc,
				PoolSize:   testPoolSize,
				LockKey:    testLockKey,
				JobOptions: JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  nil,
				WorkFunc:   workFunc,
				PoolSize:   testPoolSize,
				LockKey:    testLockKey,
				JobOptions: JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: nil,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:  []string{},
				WorkFunc:   workFunc,
				PoolSize:   testPoolSize,
				LockKey:    testLockKey,
				JobOptions: JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: []string{},
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
		},

		// Passing A Retrieval Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: retrieveFunc,
				JobOptions:   JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(1),
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: retrieveFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
			},
		},

		// Passing A Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				PoolSize:          testPoolSize,
				LockKey:           testLockKey,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions:        JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems: input,
				WorkFunc:  workFunc,
				PoolSize:  testPoolSize,
				LockKey:   testLockKey,
				// Won't cause an error because it won't be called.
				RemoveStorageFunc: errorRemoveStorageFunc,
				JobOptions:        JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
			},
		},

		// Passing Retrieval And Storage Removal Functions
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: retrieveFunc,
				// Storage removal is not called.
				RemoveStorageFunc: removeStorageFunc,
				JobOptions:        JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(1),
				WorkErrors:     map[int]error{},
			},
		},
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				PoolSize:          testPoolSize,
				LockKey:           testLockKey,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
			},
			resetStorage:      true,
			checkEmptyStorage: true,
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFuncWithStorage,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: retrieveFunc,
				// Storage removal is not called.
				RemoveStorageFunc: removeStorageFunc,
				JobOptions:        JobOptions{},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{1, 2},
				RemainingItems: []string{"CCC"},
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				// First run will store results, so second run will not have a run time.
				RunTime:    int64(0),
				WorkErrors: map[int]error{},
			},
			resetStorage: true,
		},
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFuncWithStorage,
				PoolSize:          testPoolSize,
				LockKey:           testLockKey,
				RetrieveFunc:      retrieveFunc,
				RemoveStorageFunc: removeStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
				},
			},
			initialOutput: &JobOutput[string, int]{
				ResultItems:    []int{},
				RemainingItems: input,
				RunTime:        int64(0),
				WorkErrors:     map[int]error{},
			},
			finalOutput: &JobOutput[string, int]{
				ResultItems:    finalExpected,
				RemainingItems: []string{},
				RunTime:        int64(len(input)),
				WorkErrors:     map[int]error{},
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
			errorSubstring: "Job cannot have a nil work function.",
		},
		{
			job: Job[string, int]{
				WorkItems: nil,
				WorkFunc:  nil,
			},
			errorSubstring: "Job cannot have a nil work function.",
		},

		// Bad Storage Function
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: errorRetrieveFunc,
				JobOptions:   JobOptions{},
			},
			errorSubstring: "Crazy retrieval error!",
		},
		{
			job: Job[string, int]{
				WorkItems:    input,
				WorkFunc:     workFunc,
				PoolSize:     testPoolSize,
				LockKey:      testLockKey,
				RetrieveFunc: errorRetrieveFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
				},
			},
			// Won't cause an initial error because it won't be called.
			errorSubstring: "Crazy retrieval error!",
		},

		// Bad Storage Removal Function
		{
			job: Job[string, int]{
				WorkItems:         input,
				WorkFunc:          workFunc,
				PoolSize:          testPoolSize,
				LockKey:           testLockKey,
				RemoveStorageFunc: errorRemoveStorageFunc,
				JobOptions: JobOptions{
					OverwriteRecords: true,
				},
			},
			errorSubstring: "Insane storage removal error!",
		},
	}

	for i, testCase := range testCases {
		testCase.job.WaitForCompletion = false

		output := testCase.job.Run()
		if output.Error != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(output.Error.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output on initial run. Expected substring: '%s', actual error: '%v'.", i, testCase.errorSubstring, output.Error)
				}
			} else {
				test.Errorf("Case %d: Failed to run initial job: '%v'.", i, output.Error)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected initial error: '%s'.", i, testCase.errorSubstring)
			continue
		}

		// Set the done channel to pass the equality check.
		testCase.initialOutput.Done = output.Done

		if !reflect.DeepEqual(output, testCase.initialOutput) {
			test.Errorf("Case %d: Unexpected initial results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.initialOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
			continue
		}

		testCase.job.WaitForCompletion = true
		<-output.Done

		output = testCase.job.Run()
		if output.Error != nil {
			test.Errorf("Case %d: Failed to run final job: '%v'.", i, output.Error)
			continue
		}

		// Set the done channel to pass the equality check.
		testCase.finalOutput.Done = output.Done

		if !reflect.DeepEqual(output, testCase.finalOutput) {
			test.Errorf("Case %d: Unexpected final results. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalOutput.toPrintableJobOutput()), util.MustToJSONIndent(output.toPrintableJobOutput()))
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

				output = testCase.job.Run()
				if output.Error != nil {
					test.Errorf("Case %d: Failed to check for an empty storage: '%v'.", i, output.Error)
					storage = resetStorage()
					continue
				}

				expected := &JobOutput[string, int]{
					ResultItems:    []int{},
					RemainingItems: input,
					RunTime:        int64(0),
					WorkErrors:     map[int]error{},
					// Set the done channel to pass the equality check.
					Done: output.Done,
				}

				if !reflect.DeepEqual(output, expected) {
					test.Errorf("Case %d: Unexpected output during storage check. Expected: '%v', actual: '%v'.",
						i, expected.toPrintableJobOutput(), output.toPrintableJobOutput())
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
		JobOptions: JobOptions{
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
		WorkItems: input,
		WorkFunc:  workFunc,
		PoolSize:  testPoolSize,
		LockKey:   testLockKey,
		JobOptions: JobOptions{
			WaitForCompletion: false,
		},
	}

	output := job.Run()
	if output.Error != nil {
		test.Fatalf("Failed to run job: '%v'.", output.Error)
	}

	// Wait for the worker to signal the job is done.
	<-output.Done

	expected := JobOutput[string, int]{
		ResultItems:    []int{1, 2, 3},
		RemainingItems: []string{},
		RunTime:        int64(len(input)),
		WorkErrors:     map[int]error{},
		Done:           job.Done,
	}

	// Must check the job object itself for updates.
	// The output variable is returned before the work is done.
	if !reflect.DeepEqual(job.JobOutput, expected) {
		test.Fatalf("Unexpected output. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected.toPrintableJobOutput()), util.MustToJSONIndent(job.JobOutput.toPrintableJobOutput()))
	}
}

func resetStorage() map[string]int {
	return map[string]int{
		"A":  1,
		"BB": 2,
	}
}

package util

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestRunParallelPoolMapBase(test *testing.T) {
	testCases := []struct {
		numThreads int
		hasError   bool
	}{
		{-1, true},
		{0, true},
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{10, false},
	}

	input := []string{
		"A",
		"BB",
		"CCC",
	}

	expected := PoolResult[int]{
		Results: []int{
			1,
			2,
			3,
		},
		WorkErrors:     map[int]error{},
		Canceled:       false,
		CompletedItems: []bool{true, true, true},
	}

	workFunc := func(input string) (int, error) {
		return len(input), nil
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		output, err := RunParallelPoolMap(testCase.numThreads, input, context.Background(), false, workFunc)
		if err != nil {
			if !testCase.hasError {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			}

			continue
		}

		if testCase.hasError {
			test.Errorf("Case %d: Did not get an expected error.", i)
			continue
		}

		output.Done()

		// Clear output done function for comparison check.
		output.doneFunc = nil

		if !reflect.DeepEqual(expected, output) {
			test.Errorf("Case %d: Result not as expected. Expected: '%s', actual: '%s'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		time.Sleep(25 * time.Millisecond)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}

func TestRunParallelPoolMapCancel(test *testing.T) {
	testCases := []struct {
		numThreads int
	}{
		{1},
		{2},
		{3},
		{4},
		{10},
	}

	input := []string{
		"A",
		"BB",
		"CCC",
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		// Block until the first worker has started.
		workWaitGroup := sync.WaitGroup{}
		workWaitGroup.Add(1)

		workFunc := func(input string) (int, error) {
			// Allow the first input to return a result to test partial results.
			if input == "A" {
				return len(input), nil
			}

			// Sleep to allow the first result to be captured.
			time.Sleep(2 * time.Millisecond)

			// Signal on the second piece of work so that we can make sure the workers have started up before we cancel.
			if input == "BB" {
				workWaitGroup.Done()
			}

			// Sleep to allow the the stop signal to go through.
			// We do not want to capture any more results.
			time.Sleep(2 * time.Millisecond)

			return len(input), nil
		}

		ctx, cancelFunc := context.WithCancel(context.Background())

		// Cancel the context as soon as the first worker signals it.
		go func() {
			workWaitGroup.Wait()
			cancelFunc()
		}()

		output, err := RunParallelPoolMap(testCase.numThreads, input, ctx, false, workFunc)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			continue
		}

		output.Done()

		expected := PoolResult[int]{
			Results:        []int{1, 0, 0},
			WorkErrors:     map[int]error{},
			CompletedItems: []bool{true, false, false},
			Canceled:       true,
		}

		// Clear output done function for comparison check.
		output.doneFunc = nil

		if !reflect.DeepEqual(output, expected) {
			test.Errorf("Case %d: Unexpected results. Expected: '%v', actual: '%v'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		time.Sleep(1 * time.Second)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}

func TestRunParallelPoolMapCancelSoft(test *testing.T) {
	testCases := []struct {
		numThreads int
	}{
		{1},
		{2},
		{3},
		{4},
		{10},
	}

	input := []string{
		"A",
		"BB",
		"CCC",
	}

	for i, testCase := range testCases {
		// Count the number of active threads before running.
		startThreadCount := runtime.NumGoroutine()

		// A channel to know which jobs started.
		startedChan := make(chan string, len(input))

		// Block until the first worker has started.
		workWaitGroup := sync.WaitGroup{}
		workWaitGroup.Add(1)

		workFunc := func(input string) (int, error) {
			startedChan <- input

			// Signal on the first piece of work so that we can make sure the workers have started up before we cancel.
			if input == "A" {
				workWaitGroup.Done()
				time.Sleep(time.Duration(5) * time.Millisecond)
				return len(input), nil
			}

			time.Sleep(time.Duration(5) * time.Millisecond)

			return len(input), nil
		}

		ctx, cancelFunc := context.WithCancel(context.Background())

		// Cancel the context as soon as the first worker signals it.
		go func() {
			workWaitGroup.Wait()
			cancelFunc()
		}()

		output, err := RunParallelPoolMap(testCase.numThreads, input, ctx, true, workFunc)
		if err != nil {
			test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
			continue
		}

		output.Done()

		close(startedChan)

		actualStarted := map[string]int{}
		for startedInput := range startedChan {
			actualStarted[startedInput] = len(startedInput)
		}

		// Workers race to start before the cancellation.
		// Ensure all started work is completed.
		variableExpected := make([]int, len(input))
		variableCompleted := make([]bool, len(input))
		for input, output := range actualStarted {
			switch input {
			case "A":
				variableExpected[0] = output
				variableCompleted[0] = true
			case "BB":
				variableExpected[1] = output
				variableCompleted[1] = true
			case "CCC":
				variableExpected[2] = output
				variableCompleted[2] = true
			default:
				test.Errorf("Case %d: Unexpected work input: '%s'.", i, input)
			}
		}

		expected := PoolResult[int]{
			Results:        variableExpected,
			WorkErrors:     map[int]error{},
			Canceled:       true,
			CompletedItems: variableCompleted,
		}

		// Clear output done function for comparison check.
		output.doneFunc = nil

		if !reflect.DeepEqual(output, expected) {
			test.Errorf("Case %d: Unexpected results. Expected: '%v', actual: '%v'.",
				i, MustToJSONIndent(expected), MustToJSONIndent(output))
			continue
		}

		// Check for the thread count last (this gives the workers a small bit of extra time to exit).
		// Note that there may be other tests with stray threads, so we are allowed to have less than when we started.
		// time.Sleep(25 * time.Millisecond)
		endThreadCount := runtime.NumGoroutine()
		if startThreadCount < endThreadCount {
			test.Errorf("Case %d: Ended with more threads than we started with. Start: %d, End: %d.",
				i, startThreadCount, endThreadCount)
			continue
		}
	}
}

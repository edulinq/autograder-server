package util

import (
	"testing"
)

func TestGetAllStackTracesBase(test *testing.T) {
	stacks := GetAllStackTraces()

	// Minimum stacks: test runner, this test.
	minStacks := 2
	if len(stacks) < minStacks {
		test.Fatalf("Only found %d stacks, expecting at least %d.", len(stacks), minStacks)
	}
}

func TestGetCurrentStackTraceBase(test *testing.T) {
	stack := GetCurrentStackTrace()
	if stack == nil {
		test.Fatalf("Failed to get stack trace.")
	}

	// Minimum frames: test runner, this, stack.go.
	minFrames := 3
	if len(stack.Records) < minFrames {
		test.Fatalf("Only found %d frames, expecting at least %d.", len(stack.Records), minFrames)
	}
}

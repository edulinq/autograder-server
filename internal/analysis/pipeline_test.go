package analysis

import (
	"fmt"
	"sync"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

// TestPipelineInitClose verifies that InitPipeline and ClosePipeline are idempotent
// and leave the singleton in a consistent state.
func TestPipelineInitClose(test *testing.T) {
	defer ClosePipeline()

	// Two consecutive inits should not panic or start duplicate workers.
	InitPipeline()
	InitPipeline()

	stats := GetPipelineStats()
	if stats == nil {
		test.Fatal("Expected non-nil stats after InitPipeline.")
	}

	ClosePipeline()

	stats = GetPipelineStats()
	if stats != nil {
		test.Fatal("Expected nil stats after ClosePipeline.")
	}

	// Second close should not panic.
	ClosePipeline()

	// Re-init should work cleanly after a full close.
	InitPipeline()
	stats = GetPipelineStats()
	if stats == nil {
		test.Fatal("Expected non-nil stats after second InitPipeline.")
	}
}

// TestPipelineEnqueueWhenNotRunning verifies that EnqueueSubmission is safe to call
// before InitPipeline or after ClosePipeline. The grading path calls this unconditionally,
// so it must never panic regardless of pipeline state.
func TestPipelineEnqueueWhenNotRunning(test *testing.T) {
	ClosePipeline()

	info := &model.GradingInfo{
		ID:           "course101::hw0::test@test.edulinq.org::0001",
		CourseID:     "course101",
		AssignmentID: "hw0",
		User:         "test@test.edulinq.org",
		Score:        1.0,
	}

	// Must not panic.
	EnqueueSubmission(info)
}

// setTestPipeline replaces the global pipeline with a test instance that has no workers
// (so jobs sit in the channel for inspection) and returns a cleanup function that restores
// the previous global. The returned pipeline uses a nil ctx because workers are not started
// and nothing calls p.ctx directly on the ingest path.
func setTestPipeline(queueDepth int) (*Pipeline, func()) {
	p := &Pipeline{
		queue: make(chan pipelineJob, queueDepth),
	}

	pipelineMu.Lock()
	old := globalPipeline
	globalPipeline = p
	pipelineMu.Unlock()

	return p, func() {
		pipelineMu.Lock()
		globalPipeline = old
		pipelineMu.Unlock()
	}
}

// TestPipelineDropsWhenFull verifies the backpressure behavior: when the queue is at
// capacity, additional EnqueueSubmission calls are dropped without blocking.
//
// Run with -race. The race condition this guards against: an earlier draft had a
// blocking channel send instead of the non-blocking select. Under concurrent load,
// goroutines would serialize waiting for a worker to drain the channel, coupling
// grading latency to analysis latency. The non-blocking select + atomic counter
// approach eliminates that coupling.
func TestPipelineDropsWhenFull(test *testing.T) {
	const queueDepth = 3

	_, restore := setTestPipeline(queueDepth)
	defer restore()

	makeInfo := func(i int) *model.GradingInfo {
		return &model.GradingInfo{
			ID:           fmt.Sprintf("course101::hw0::test@test.edulinq.org::%04d", i),
			CourseID:     "course101",
			AssignmentID: "hw0",
			User:         "test@test.edulinq.org",
			Score:        float64(i),
		}
	}

	const totalJobs = 20

	var wg sync.WaitGroup
	for i := 0; i < totalJobs; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			EnqueueSubmission(makeInfo(idx))
		}(i)
	}

	wg.Wait()

	stats := GetPipelineStats()
	if stats == nil {
		test.Fatal("Expected non-nil stats.")
	}

	// Exactly queueDepth jobs should have made it into the channel.
	if stats.Enqueued != int64(queueDepth) {
		test.Errorf("Expected %d enqueued (queue capacity), got %d.", queueDepth, stats.Enqueued)
	}

	if stats.Dropped != int64(totalJobs-queueDepth) {
		test.Errorf("Expected %d dropped, got %d.", totalJobs-queueDepth, stats.Dropped)
	}

	if stats.QueueLen != queueDepth {
		test.Errorf("Expected queue length %d, got %d.", queueDepth, stats.QueueLen)
	}
}

// TestPipelineEnqueuedAtTimestamp verifies that the enqueuedAt field is set before the
// job enters the channel. This matters for lag calculation: if the timestamp were set
// after a blocking send completed, it would undercount the time spent waiting.
// In the non-blocking select design this is a non-issue, but this test makes the
// ordering invariant explicit and will catch regressions if the send path changes.
func TestPipelineEnqueuedAtTimestamp(test *testing.T) {
	_, restore := setTestPipeline(10)
	defer restore()

	before := timestamp.Now()

	EnqueueSubmission(&model.GradingInfo{
		ID:           "course101::hw0::test@test.edulinq.org::0001",
		CourseID:     "course101",
		AssignmentID: "hw0",
		User:         "test@test.edulinq.org",
	})

	after := timestamp.Now()

	pipelineMu.Lock()
	p := globalPipeline
	pipelineMu.Unlock()

	job := <-p.queue

	if job.enqueuedAt < before || job.enqueuedAt > after {
		test.Errorf("enqueuedAt %v is outside expected range [%v, %v].",
			job.enqueuedAt, before, after)
	}
}

// TestPipelineConcurrentInitClose verifies that concurrent InitPipeline and ClosePipeline
// calls do not race on the globalPipeline pointer. The -race detector will catch missing
// or incorrectly scoped mutex usage. This is the test that would have caught the bug in
// an earlier design that dropped the mutex lock before setting globalPipeline = nil.
func TestPipelineConcurrentInitClose(test *testing.T) {
	defer ClosePipeline()

	const goroutines = 20

	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			InitPipeline()
		}()
	}

	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ClosePipeline()
		}()
	}

	wg.Wait()
}

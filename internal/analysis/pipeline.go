package analysis

// pipeline.go - Background async analysis pipeline for post-grading feature extraction.
//
// This is deliberately separate from the request-driven IndividualAnalysis and PairwiseAnalysis
// paths. Those paths are synchronous from the caller's perspective and tied to specific API
// requests with contexts that get canceled when HTTP connections close. This pipeline runs
// continuously in the background, processing grading events without any coupling to grading
// latency or HTTP request lifecycles.
//
// The core contract: EnqueueSubmission() must never block. If the queue is full, the job is
// dropped and logged. Some submissions may not have pipeline analysis results; that is normal
// and expected, not an error condition.

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

// pipelineJob is small on purpose. Workers fetch the full submission from the DB;
// we don't hold grading result data in memory across the queue.
type pipelineJob struct {
	fullSubmissionID string
	courseID         string
	assignmentID     string
	userEmail        string
	enqueuedAt       timestamp.Timestamp
}

// Pipeline manages a fixed-size pool of workers draining a buffered job channel.
// The zero value is not usable; construct via InitPipeline().
type Pipeline struct {
	queue  chan pipelineJob
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Atomic counters for observability. No locks needed — these are strictly additive.
	enqueued  atomic.Int64
	dropped   atomic.Int64
	processed atomic.Int64
	failed    atomic.Int64
}

// Global singleton. nil means the pipeline is not running.
// All access to globalPipeline must hold pipelineMu.
var (
	globalPipeline *Pipeline
	pipelineMu     sync.Mutex
)

// InitPipeline starts the background analysis pipeline. Safe to call multiple times;
// subsequent calls are no-ops if the pipeline is already running.
// Call this at server startup, before any grading requests are served.
func InitPipeline() {
	pipelineMu.Lock()
	defer pipelineMu.Unlock()

	if globalPipeline != nil {
		return
	}

	depth := config.ANALYSIS_PIPELINE_QUEUE_DEPTH.Get()
	workers := config.ANALYSIS_PIPELINE_WORKERS.Get()

	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &Pipeline{
		queue:  make(chan pipelineJob, depth),
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}

	globalPipeline = p

	log.Info("Analysis pipeline started.",
		log.NewAttr("workers", workers),
		log.NewAttr("queue-depth", depth),
	)
}

// ClosePipeline signals all workers to stop and waits for in-flight jobs to finish.
// After this returns, the pipeline may be re-initialized with InitPipeline().
// Intended for clean server shutdown and test teardown.
func ClosePipeline() {
	pipelineMu.Lock()
	p := globalPipeline
	globalPipeline = nil
	pipelineMu.Unlock()

	if p == nil {
		return
	}

	// Cancel signals workers to stop picking up new jobs.
	// Workers currently in processJob() will finish their current job before exiting.
	p.cancel()
	p.wg.Wait()
}

// EnqueueSubmission submits a completed grading event for async feature extraction.
// This is the integration point called from internal/grader/grade.go after db.SaveSubmission
// succeeds. The call must never block — if the pipeline queue is full, the job is dropped.
func EnqueueSubmission(info *model.GradingInfo) {
	pipelineMu.Lock()
	p := globalPipeline
	pipelineMu.Unlock()

	if p == nil {
		// Pipeline not running. This is normal in unit tests and during shutdown.
		return
	}

	job := pipelineJob{
		fullSubmissionID: info.ID,
		courseID:         info.CourseID,
		assignmentID:     info.AssignmentID,
		userEmail:        info.User,
		enqueuedAt:       timestamp.Now(),
	}

	select {
	case p.queue <- job:
		p.enqueued.Add(1)
	default:
		// The channel is full. Drop rather than block. The grading result is already
		// saved to the DB, so the student's grade is not affected. The pipeline analysis
		// for this submission simply won't happen.
		p.dropped.Add(1)

		log.Warn("Analysis pipeline queue full, dropping submission.",
			log.NewCourseAttr(info.CourseID),
			log.NewAssignmentAttr(info.AssignmentID),
			log.NewUserAttr(info.User),
			log.NewAttr("queue-depth", len(p.queue)),
		)
	}
}

func (p *Pipeline) runWorker() {
	defer p.wg.Done()

	// Each worker gets its own RNG so there are no races on the random source.
	// rand.Rand is not goroutine-safe; a shared instance would require a mutex
	// and would serialize noise generation across workers.
	rng := newPrivatizerRNG()

	for {
		select {
		case <-p.ctx.Done():
			// Drain any remaining jobs before exiting so we don't silently discard
			// work that was already enqueued before shutdown was signaled.
			p.drainRemaining(rng)
			return

		case job, ok := <-p.queue:
			if !ok {
				return
			}

			p.processJob(job, rng)
		}
	}
}

// drainRemaining processes any jobs already in the queue after the context is canceled.
// We do a best-effort drain rather than abandoning them, since these jobs represent
// submissions that are fully graded and saved — the extraction work is cheap relative
// to the cost of a lost analysis result.
func (p *Pipeline) drainRemaining(rng *rand.Rand) {
	for {
		select {
		case job, ok := <-p.queue:
			if !ok {
				return
			}

			p.processJob(job, rng)

		default:
			// Nothing left.
			return
		}
	}
}

func (p *Pipeline) processJob(job pipelineJob, rng *rand.Rand) {
	// Build an AnalysisOptions that mirrors what IndividualAnalysis would use for a
	// single submission. We set WaitForCompletion because we're already in a background
	// goroutine — there's no reason to spawn further goroutines inside the job.
	// RetainOriginalContext is true so the analysis uses p.ctx rather than swapping in
	// context.Background() (the analysis package's default for async calls).
	opts := AnalysisOptions{
		ResolvedSubmissionIDs: []string{job.fullSubmissionID},
		RetainOriginalContext: true,
		JobOptions: jobmanager.JobOptions{
			WaitForCompletion: true,
			Context:           p.ctx,
		},
	}

	// computeSingleIndividualAnalysis handles temp dir creation and cleanup internally.
	// We pass computeDeltas=false because the delta computation requires fetching the
	// previous submission and running analysis on it too — that doubles the work per
	// job and the delta values will be available via the standard IndividualAnalysis
	// path anyway. For the DP pipeline we only need point-in-time features.
	result, err := computeSingleIndividualAnalysis(opts, job.fullSubmissionID, false)
	if err != nil {
		p.failed.Add(1)

		log.Error("Pipeline failed to analyze submission.",
			err,
			log.NewCourseAttr(job.courseID),
			log.NewAssignmentAttr(job.assignmentID),
			log.NewUserAttr(job.userEmail),
		)

		return
	}

	// result == nil means the context was canceled mid-analysis.
	if result == nil {
		return
	}

	// A non-nil result with Failure=true means the analysis ran but encountered a
	// recoverable problem (e.g., unsupported file type). Log it but don't retry —
	// retrying wouldn't help since the same submission would produce the same result.
	if result.Failure {
		p.failed.Add(1)

		log.Warn("Pipeline analysis produced a failure result.",
			log.NewCourseAttr(job.courseID),
			log.NewAssignmentAttr(job.assignmentID),
			log.NewUserAttr(job.userEmail),
			log.NewAttr("reason", result.FailureMessage),
		)

		return
	}

	p.processed.Add(1)

	// Warn if we're falling behind. A queue depth >10% of capacity sustained over
	// multiple jobs suggests the worker count or queue depth needs tuning.
	queueLen := len(p.queue)
	queueCap := cap(p.queue)
	if queueCap > 0 && queueLen > queueCap/10 {
		lagMS := (timestamp.Now() - job.enqueuedAt).ToMSecs()

		log.Warn("Analysis pipeline processing lag.",
			log.NewAttr("queue-len", queueLen),
			log.NewAttr("queue-cap", queueCap),
			log.NewAttr("lag-ms", lagMS),
		)
	}

	// Apply differential privacy and persist the sanitized result.
	// Failures here are best-effort: log and move on. The student's grade is
	// already saved; this is research data collection only.
	cfg := PrivacyConfigFromGlobalConfig()

	err = privatizeAndStore(result, cfg, rng)
	if err != nil {
		logPrivacyError(err, job.fullSubmissionID)
	}
}

// PipelineStats is a snapshot of pipeline activity counters.
// Enqueued = Dropped + Processed + Failed + currently in-flight.
type PipelineStats struct {
	QueueLen  int
	QueueCap  int
	Enqueued  int64
	Dropped   int64
	Processed int64
	Failed    int64
}

// GetPipelineStats returns a snapshot of pipeline counters, or nil if the pipeline
// is not running. Useful for health-check endpoints and integration tests.
func GetPipelineStats() *PipelineStats {
	pipelineMu.Lock()
	p := globalPipeline
	pipelineMu.Unlock()

	if p == nil {
		return nil
	}

	return &PipelineStats{
		QueueLen:  len(p.queue),
		QueueCap:  cap(p.queue),
		Enqueued:  p.enqueued.Load(),
		Dropped:   p.dropped.Load(),
		Processed: p.processed.Load(),
		Failed:    p.failed.Load(),
	}
}

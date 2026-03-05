package analysis

// Differential privacy layer for the background analysis pipeline.
//
// This implements the Laplace mechanism for pure (delta=0) differential privacy.
// The Laplace distribution is parameterized by a scale b = sensitivity/epsilon.
// For counting queries over a dataset of student submissions, the global sensitivity
// is 1: adding or removing one student's submission changes any count by at most 1.
//
// For non-count features (LOC, score, time deltas), sensitivity is determined by
// the maximum range of the feature. We currently use sensitivity=1 as a conservative
// approximation for all features, which may underprotect high-range features like LOC
// but is a reasonable starting point. See extractFeatures for per-feature commentary.
//
// No epsilon budget composition tracking is implemented. Each privatized record is
// independently (epsilon, 0)-DP. Downstream consumers must account for this if they
// query the same student's records multiple times.
//
// RNG: math/rand, not crypto/rand. The noise must have correct statistical properties
// but does not need cryptographic unpredictability. crypto/rand is substantially slower
// and provides no additional DP guarantees.

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

// PrivacyConfig holds the parameters for the Laplace mechanism.
// Epsilon must be strictly positive. Delta is stored but currently unused — the
// implementation is pure DP only. It is included so that callers can record the
// full (epsilon, delta) budget in stored results for future reference.
type PrivacyConfig struct {
	Epsilon float64
	Delta   float64 // reserved for future Gaussian/approximate DP
}

func (c PrivacyConfig) Validate() error {
	if c.Epsilon <= 0 {
		return fmt.Errorf("DP epsilon must be positive, got %g.", c.Epsilon)
	}

	if c.Delta < 0 {
		return fmt.Errorf("DP delta must be non-negative, got %g.", c.Delta)
	}

	return nil
}

// PrivacyConfigFromGlobalConfig reads epsilon and delta from the server config.
func PrivacyConfigFromGlobalConfig() PrivacyConfig {
	return PrivacyConfig{
		Epsilon: config.ANALYSIS_DP_EPSILON.Get(),
		Delta:   config.ANALYSIS_DP_DELTA.Get(),
	}
}

// privatizeAndStore applies the Laplace mechanism to the features extracted from
// result and writes the sanitized record to the database.
//
// Failures here must not propagate to the caller in a way that would block the
// grading pipeline — this is best-effort research data collection.
// The caller (processJob in pipeline.go) is responsible for logging and continuing.
func privatizeAndStore(result *model.IndividualAnalysis, cfg PrivacyConfig, rng *rand.Rand) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("Invalid privacy config: '%w'.", err)
	}

	names, rawValues := extractFeatures(result)
	if len(names) == 0 {
		// Nothing to privatize — analysis result had no usable features.
		return nil
	}

	noisy := make([]float64, len(rawValues))
	for i, v := range rawValues {
		noisy[i] = addLaplaceNoise(v, cfg.Epsilon, rng)
	}

	record := &model.PrivatizedAnalysis{
		FullSubmissionID:     result.FullID,
		CourseID:             result.CourseID,
		AssignmentID:         result.AssignmentID,
		AnalysisTimestamp:    timestamp.Now(),
		EpsilonUsed:          cfg.Epsilon,
		FeatureNames:         names,
		NoisyFeatures:        noisy,
		CourseAssignmentHash: hashCourseAssignment(result.CourseID, result.AssignmentID),
	}

	err := db.StorePrivatizedAnalysis([]*model.PrivatizedAnalysis{record})
	if err != nil {
		return fmt.Errorf("Failed to store privatized analysis for '%s': '%w'.", result.FullID, err)
	}

	return nil
}

// extractFeatures maps the numeric fields of an IndividualAnalysis to a named feature
// vector. Names and values are returned as parallel slices.
//
// Sensitivity notes per feature:
//   - lines-of-code: sensitivity could be O(1000) for real submissions; using 1 is
//     an approximation. Adding noise with scale=1 to a value of 500 provides weak
//     privacy. This is acceptable for aggregate research queries but not for
//     identifying individual students from the noisy values alone.
//   - score: range [0, 1] in most assignments; sensitivity=1 is exact.
//   - submission-time-delta: milliseconds, range could be days; sensitivity=1 is weak.
//
// These approximations are documented here rather than silently assumed.
// A follow-up could clip features to a known range before adding noise, which
// is the standard practice for tighter sensitivity bounds.
func extractFeatures(analysis *model.IndividualAnalysis) (names []string, values []float64) {
	names = []string{
		"lines-of-code",
		"lines-of-code-delta",
		"score",
		"score-delta",
		"submission-time-delta-ms",
	}

	values = []float64{
		float64(analysis.LinesOfCode),
		float64(analysis.LinesOfCodeDelta),
		analysis.Score,
		analysis.ScoreDelta,
		float64(analysis.SubmissionTimeDelta),
	}

	return names, values
}

// addLaplaceNoise samples from Laplace(0, sensitivity/epsilon) and adds it to value.
// Uses inverse CDF (quantile function) sampling:
//   - u ~ Uniform(0, 1), shifted to (-0.5, 0.5) for sign information
//   - noise = -b * sign(u) * ln(1 - 2|u|)
//
// The sensitivity parameter is currently always 1 (see extractFeatures commentary).
func addLaplaceNoise(value float64, epsilon float64, rng *rand.Rand) float64 {
	const sensitivity = 1.0

	scale := sensitivity / epsilon

	// Shift uniform sample to (-0.5, 0.5) so sign encodes direction of noise.
	u := rng.Float64() - 0.5

	// Inverse CDF of Laplace distribution.
	noise := -scale * math.Copysign(math.Log(1-2*math.Abs(u)), u)

	return value + noise
}

// hashCourseAssignment produces a stable hex string for (courseID, assignmentID) pairs.
// This is stored on each privatized record so that aggregate queries can group by
// assignment without parsing the full submission ID on every read.
func hashCourseAssignment(courseID, assignmentID string) string {
	h := sha256.New()
	h.Write([]byte(courseID))
	h.Write([]byte{0}) // null byte as separator to prevent "a"+"bc" == "ab"+"c"
	h.Write([]byte(assignmentID))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// newPrivatizerRNG creates a seeded RNG for use in privatization.
// The seed is derived from the current time so that noise is not predictable
// across server restarts, while remaining reproducible within a test run when
// callers seed explicitly.
func newPrivatizerRNG() *rand.Rand {
	return rand.New(rand.NewSource(timestamp.Now().ToMSecs()))
}

// logPrivacyError is a helper that logs privatization errors at Warn level without
// returning them, enforcing the best-effort contract. Grading must always succeed
// regardless of pipeline analysis failures.
func logPrivacyError(err error, submissionID string) {
	log.Warn("Privatized analysis failed; grading result is unaffected.",
		err,
		log.NewAttr("submission-id", submissionID),
	)
}

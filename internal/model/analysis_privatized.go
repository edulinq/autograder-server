package model

import "github.com/edulinq/autograder/internal/timestamp"

// PrivatizedAnalysis holds the differentially-private version of features extracted
// from a single submission. It is stored separately from IndividualAnalysis so that
// the raw results and the noisy results have clearly distinct storage paths and
// neither accidentally overwrites the other.
//
// Feature names and values are stored as parallel slices rather than a map so that
// the order is deterministic for tests and so that schema evolution (adding/removing
// features) doesn't silently break old records: a record with 4 features and a query
// expecting 5 just returns fewer values rather than silently zero-filling.
type PrivatizedAnalysis struct {
	FullSubmissionID string `json:"submission-id"`
	CourseID         string `json:"course-id"`
	AssignmentID     string `json:"assignment-id"`

	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`

	// The epsilon value used when this record was generated. Stored so that downstream
	// consumers can filter out records with different privacy budgets when computing
	// aggregates. No composition tracking across queries — each record is independently
	// (epsilon, 0)-DP.
	EpsilonUsed float64 `json:"epsilon-used"`

	// Parallel slices — FeatureNames[i] names NoisyFeatures[i].
	FeatureNames  []string  `json:"feature-names"`
	NoisyFeatures []float64 `json:"noisy-features"`

	// Precomputed hash of (course-id, assignment-id) for grouping aggregate queries
	// without splitting the submission ID string on every read.
	CourseAssignmentHash string `json:"course-assignment-hash"`
}

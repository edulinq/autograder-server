package core

import (
	"github.com/edulinq/autograder/internal/model"
)

type SimilarityEngine interface {
	GetName() string
	IsAvailable() bool
	// Get the similarity results between two files.
	// Working on two files (submissions) at a time will typically be less efficient than working on all files at the same time,
	// but a lot of shorter jobs is more flexible than one large job.
	// In additional to similarity, the engine should also return the time it took (in milliseconds) to run
	// (not counting time locked/waiting).
	// On an error, 0 should be returned for run time.
	ComputeFileSimilarity(paths [2]string, baseLockKey string) (*model.FileSimilarity, int64, error)
}

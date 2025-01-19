package core

import (
	"github.com/edulinq/autograder/internal/model"
)

type SimilarityEngine interface {
	GetName() string
	// Get the similiarty results between two files.
	// Working on two files (submissions) at a time will typically be less efficient than working on all files at the same time,
	// but a lot of shorter jobs is more flexible than one large job.
	ComputeFileSimilarity(paths [2]string, baseLockKey string) (*model.FileSimilarity, error)
}

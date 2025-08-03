package core

import (
	"context"

	"github.com/edulinq/autograder/internal/model"
)

type SimilarityEngine interface {
	GetName() string
	IsAvailable() bool
	// Get the similarity results between two files.
	// If the engine supports it and the templatePath is not empty,
	// ignore the code in the template file when computing similarity.
	// The engine must not modify the existing template file.
	// Working on two files (submissions) at a time will typically be less efficient than working on all files at the same time,
	// but a lot of shorter jobs is more flexible than one large job.
	// On a timeout, (nil, nil) should be returned.
	// Options carries engine‑specific parameters, e.g. "min-tokens" for JPlag.
	ComputeFileSimilarity(paths [2]string, templatePath string, ctx context.Context, options map[string]any) (*model.FileSimilarity, error)
}

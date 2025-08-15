package core

import (
	"context"
	"fmt"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
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
	// The options map carries engine‑specific parameters, e.g., the minimum number of words required to perform similarity analysis by JPlag.
	ComputeFileSimilarity(paths [2]string, templatePath string, ctx context.Context, rawOptions model.OptionsMap) (*model.FileSimilarity, error)
}

func ParseEngineOptions[T any](rawOptions model.OptionsMap, defaultOptions *T) (*T, error) {
	if rawOptions == nil {
		return defaultOptions, nil
	}

	effectiveOptions, err := util.JSONTransformTypes(rawOptions, defaultOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert raw options: '%w'.", err)
	}

	return effectiveOptions, nil
}

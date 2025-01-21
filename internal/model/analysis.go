package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// A key for pairwise analysis.
// Should always be an ordered (lexicographically) pair of full submissions IDs.
type PairwiseKey [2]string

const PAIRWISE_KEY_DELIM string = "||"

type IndividualAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`

	SubmissionStartTime timestamp.Timestamp `json:"submission-start-time"`
	Files               []string            `json:"files"`

	LinesOfCode int     `json:"lines-of-code"`
	Score       float64 `json:"score"`

	LinesOfCodeDelta float64 `json:"lines-of-code-delta"`
	ScoreDelta       float64 `json:"score-delta"`

	LinesOfCodeVelocity float64 `json:"lines-of-code-velocity"`
	ScoreVelocity       float64 `json:"score-velocity"`
}

type PairwiseAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`
	SubmissionIDs     PairwiseKey         `json:"submission-ids"`

	Similarities   map[string][]*FileSimilarity `json:"similarities"`
	UnmatchedFiles [][2]string                  `json:"unmatched-files"`

	MeanSimilarities    map[string]float64 `json:"mean-similarities"`
	TotalMeanSimilarity float64            `json:"total-mean-similarity"`
}

type FileSimilarity struct {
	Filename string         `json:"filename"`
	Tool     string         `json:"tool"`
	Version  string         `json:"version"`
	Options  map[string]any `json:"options,omitempty"`
	Score    float64        `json:"score"`
}

type PairwiseAnalysisSummary struct {
	Complete      bool `json:"complete"`
	CompleteCount int  `json:"complete-count"`
	PendingCount  int  `json:"pending-count"`

	FirstTimestamp timestamp.Timestamp `json:"first-timestamp"`
	LastTimestamp  timestamp.Timestamp `json:"last-timestamp"`

	AggregateMeanSimilarities      map[string]util.AggregateValues `json:"aggregate-mean-similarities"`
	AggregateTotalMeanSimilarities util.AggregateValues            `json:"aggregate-total-mean-similarity"`
}

func NewPairwiseKey(fullSubmissionID1 string, fullSubmissionID2 string) PairwiseKey {
	return PairwiseKey([2]string{
		min(fullSubmissionID1, fullSubmissionID2),
		max(fullSubmissionID1, fullSubmissionID2),
	})
}

func (this *PairwiseKey) String() string {
	return this[0] + PAIRWISE_KEY_DELIM + this[1]
}

func NewPairwiseAnalysis(pairwiseKey PairwiseKey, similarities map[string][]*FileSimilarity, unmatches [][2]string) *PairwiseAnalysis {
	meanSimilarities := make(map[string]float64, len(similarities))
	totalMeanSimilarity := 0.0

	for relpath, sims := range similarities {
		value := 0.0
		for _, sim := range sims {
			value += sim.Score
		}

		if len(sims) > 0 {
			value /= float64(len(sims))
		}

		meanSimilarities[relpath] = value
		totalMeanSimilarity += value
	}

	if len(similarities) > 0 {
		totalMeanSimilarity /= float64(len(similarities))
	}

	return &PairwiseAnalysis{
		AnalysisTimestamp:   timestamp.Now(),
		SubmissionIDs:       pairwiseKey,
		Similarities:        similarities,
		UnmatchedFiles:      unmatches,
		MeanSimilarities:    meanSimilarities,
		TotalMeanSimilarity: totalMeanSimilarity,
	}
}

func NewPairwiseAnalysisSummary(results []*PairwiseAnalysis, pendingCount int) *PairwiseAnalysisSummary {
	if len(results) == 0 {
		return &PairwiseAnalysisSummary{
			Complete:       (pendingCount == 0),
			CompleteCount:  0,
			PendingCount:   pendingCount,
			FirstTimestamp: timestamp.Zero(),
			LastTimestamp:  timestamp.Zero(),
		}
	}

	firstTimestamp := timestamp.Zero()
	lastTimestamp := timestamp.Zero()

	meanSims := make(map[string][]float64)
	totalMeanSim := make([]float64, 0, len(results))

	for i, result := range results {
		if (i == 0) || (result.AnalysisTimestamp < firstTimestamp) {
			firstTimestamp = result.AnalysisTimestamp
		}

		if (i == 0) || (result.AnalysisTimestamp > lastTimestamp) {
			lastTimestamp = result.AnalysisTimestamp
		}

		for relpath, meanSim := range result.MeanSimilarities {
			meanSims[relpath] = append(meanSims[relpath], meanSim)
		}

		totalMeanSim = append(totalMeanSim, result.TotalMeanSimilarity)
	}

	aggregateMeanSimilarities := make(map[string]util.AggregateValues, len(meanSims))
	for relpath, meanSimValues := range meanSims {
		aggregateMeanSimilarities[relpath] = util.ComputeAggregates(meanSimValues)
	}

	return &PairwiseAnalysisSummary{
		Complete:                       (pendingCount == 0),
		CompleteCount:                  len(results),
		PendingCount:                   pendingCount,
		FirstTimestamp:                 firstTimestamp,
		LastTimestamp:                  lastTimestamp,
		AggregateMeanSimilarities:      aggregateMeanSimilarities,
		AggregateTotalMeanSimilarities: util.ComputeAggregates(totalMeanSim),
	}
}

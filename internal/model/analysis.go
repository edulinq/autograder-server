package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// A key for pairwise analysis.
// Should always be an ordered (lexicographically) pair of full submissions IDs.
type PairwiseKey [2]string

const PAIRWISE_KEY_DELIM string = "||"

type AnalysisFileInfo struct {
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original-filename,omitempty"`
	LinesOfCode      int    `json:"lines-of-code"`
}

type FileSimilarity struct {
	AnalysisFileInfo

	Tool    string         `json:"tool"`
	Version string         `json:"version"`
	Options map[string]any `json:"options,omitempty"`
	Score   float64        `json:"score"`
}

type AnalysisSummary struct {
	Complete      bool `json:"complete"`
	CompleteCount int  `json:"complete-count"`
	PendingCount  int  `json:"pending-count"`

	FirstTimestamp timestamp.Timestamp `json:"first-timestamp"`
	LastTimestamp  timestamp.Timestamp `json:"last-timestamp"`
}

type IndividualAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`

	FullID       string `json:"submission-id"`
	ShortID      string `json:"short-id"`
	CourseID     string `json:"course-id"`
	AssignmentID string `json:"assignment-id"`
	UserEmail    string `json:"user-email"`

	SubmissionStartTime timestamp.Timestamp `json:"submission-start-time"`
	Score               float64             `json:"score"`

	Files       []AnalysisFileInfo `json:"files"`
	LinesOfCode int                `json:"lines-of-code"`

	LinesOfCodeDelta int     `json:"lines-of-code-delta"`
	ScoreDelta       float64 `json:"score-delta"`

	LinesOfCodeVelocity float64 `json:"lines-of-code-per-hour"`
	ScoreVelocity       float64 `json:"score-per-hour"`
}

type IndividualAnalysisSummary struct {
	AnalysisSummary

	AggregateScore util.AggregateValues `json:"aggregate-score"`

	AggregateLinesOfCode        util.AggregateValues            `json:"aggregate-lines-of-code"`
	AggregateLinesOfCodePerFile map[string]util.AggregateValues `json:"aggregate-lines-of-code-per-file"`

	AggregateLinesOfCodeDelta util.AggregateValues `json:"aggregate-lines-of-code-delta"`
	AggregateScoreDelta       util.AggregateValues `json:"aggregate-score-delta"`

	AggregateLinesOfCodeVelocity util.AggregateValues `json:"aggregate-lines-of-code-per-hour"`
	AggregateScoreVelocity       util.AggregateValues `json:"aggregate-score-per-hour"`
}

type PairwiseAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`
	SubmissionIDs     PairwiseKey         `json:"submission-ids"`

	Similarities   map[string][]*FileSimilarity `json:"similarities"`
	UnmatchedFiles [][2]string                  `json:"unmatched-files"`

	MeanSimilarities    map[string]float64 `json:"mean-similarities"`
	TotalMeanSimilarity float64            `json:"total-mean-similarity"`
}

type PairwiseAnalysisSummary struct {
	AnalysisSummary

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

func NewIndividualAnalysisSummary(results []*IndividualAnalysis, pendingCount int) *IndividualAnalysisSummary {
	if len(results) == 0 {
		return &IndividualAnalysisSummary{
			AnalysisSummary: AnalysisSummary{
				Complete:       (pendingCount == 0),
				CompleteCount:  0,
				PendingCount:   pendingCount,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
		}
	}

	firstTimestamp := timestamp.Zero()
	lastTimestamp := timestamp.Zero()

	scores := make([]float64, 0, len(results))
	locs := make([]float64, 0, len(results))
	locDeltas := make([]float64, 0, len(results))
	scoreDeltas := make([]float64, 0, len(results))
	locVelocities := make([]float64, 0, len(results))
	scoreVelocities := make([]float64, 0, len(results))

	locPerFiles := make(map[string][]float64)

	for i, result := range results {
		if (i == 0) || (result.AnalysisTimestamp < firstTimestamp) {
			firstTimestamp = result.AnalysisTimestamp
		}

		if (i == 0) || (result.AnalysisTimestamp > lastTimestamp) {
			lastTimestamp = result.AnalysisTimestamp
		}

		for _, info := range result.Files {
			locPerFiles[info.Filename] = append(locPerFiles[info.Filename], float64(info.LinesOfCode))
		}

		scores = append(scores, result.Score)
		locs = append(scores, float64(result.LinesOfCode))
		locDeltas = append(scores, float64(result.LinesOfCodeDelta))
		scoreDeltas = append(scores, result.ScoreDelta)
		locVelocities = append(scores, result.LinesOfCodeVelocity)
		scoreVelocities = append(scores, result.ScoreVelocity)
	}

	aggregateLOCPerFile := make(map[string]util.AggregateValues, len(locPerFiles))
	for relpath, locValues := range locPerFiles {
		aggregateLOCPerFile[relpath] = util.ComputeAggregates(locValues)
	}

	return &IndividualAnalysisSummary{
		AnalysisSummary: AnalysisSummary{
			Complete:       (pendingCount == 0),
			CompleteCount:  len(results),
			PendingCount:   pendingCount,
			FirstTimestamp: firstTimestamp,
			LastTimestamp:  lastTimestamp,
		},
		AggregateScore:               util.ComputeAggregates(scores),
		AggregateLinesOfCode:         util.ComputeAggregates(locs),
		AggregateLinesOfCodePerFile:  aggregateLOCPerFile,
		AggregateLinesOfCodeDelta:    util.ComputeAggregates(locDeltas),
		AggregateScoreDelta:          util.ComputeAggregates(scoreDeltas),
		AggregateLinesOfCodeVelocity: util.ComputeAggregates(locVelocities),
		AggregateScoreVelocity:       util.ComputeAggregates(scoreVelocities),
	}
}

func NewPairwiseAnalysisSummary(results []*PairwiseAnalysis, pendingCount int) *PairwiseAnalysisSummary {
	if len(results) == 0 {
		return &PairwiseAnalysisSummary{
			AnalysisSummary: AnalysisSummary{
				Complete:       (pendingCount == 0),
				CompleteCount:  0,
				PendingCount:   pendingCount,
				FirstTimestamp: timestamp.Zero(),
				LastTimestamp:  timestamp.Zero(),
			},
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
		AnalysisSummary: AnalysisSummary{
			Complete:       (pendingCount == 0),
			CompleteCount:  len(results),
			PendingCount:   pendingCount,
			FirstTimestamp: firstTimestamp,
			LastTimestamp:  lastTimestamp,
		},
		AggregateMeanSimilarities:      aggregateMeanSimilarities,
		AggregateTotalMeanSimilarities: util.ComputeAggregates(totalMeanSim),
	}
}

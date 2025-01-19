package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

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

type PairWiseAnalysis struct {
	AnalysisTimestamp timestamp.Timestamp `json:"analysis-timestamp"`
	SubmissionIDs     [2]string           `json:"submission-ids"`

	Similarities   map[string][]*FileSimilarity `json:"similarities"`
	UnmatchedFiles [][2]string                  `json:"unmatched-files"`
}

type FileSimilarity struct {
	Filename string         `json:"filename"`
	Tool     string         `json:"tool"`
	Version  string         `json:"version"`
	Options  map[string]any `json:"options,omitempty"`
	Score    float64        `json:"score"`
}

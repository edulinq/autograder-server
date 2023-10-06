package model

import (
    "time"
)

const SCORING_INFO_IDENTITY_KEY string = "__autograder__";

type ScoringInfo struct {
    ID string `json:"id"`
    SubmissionTime time.Time `json:"submission-time"`
    UploadTime time.Time `json:"upload-time"`
    RawScore float64 `json:"raw-score"`
    Score float64 `json:"score"`
    Lock bool `json:"lock"`
    LateDayUsage int `json:"late-date-usage"`
    NumDaysLate int `json:"num-days-late"`
    Reject bool `json:"reject"`

    // A distinct key so we can recognize this as an aautograder object.
    Autograder int `json:"__autograder__"`
    // If this object was serialized from a Canvas comment, keep the ID.
    CanvasCommentID string `json:"-"`
}

func ScoringInfoFromSubmissionSummary(summary *SubmissionSummary) *ScoringInfo {
    return &ScoringInfo{
        ID: summary.ID,
        SubmissionTime: summary.GradingStartTime,
        RawScore: summary.Score,
    };
}

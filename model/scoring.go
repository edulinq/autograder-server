package model

import (
    "time"
)

type ScoringInfo struct {
    ID string `json:"id"`
    SubmissionTime time.Time `json:"submission-time"`
    UploadTime time.Time `json:"upload-time"`
    RawScore float64 `json:"raw-score"`
    Score float64 `json:"score"`
    Lock bool `json:"lock"`
    LateDayUsage int `json:"late-date-usage"`
    NumDaysLate int `json:"num-days-late"`
    Rejected bool `json:"rejected"`
}

func ScoringInfoFromSubmissionSummary(summary *SubmissionSummary) *ScoringInfo {
    return &ScoringInfo{
        ID: summary.ID,
        SubmissionTime: summary.GradingStartTime,
        RawScore: summary.Score,
    };
}

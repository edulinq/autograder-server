package artifact

import (
    "time"

    "github.com/eriq-augustine/autograder/util"
)

type TestSubmission struct {
    IgnoreMessages bool `json:"ignore_messages"`
    Result GradedAssignment `json:"result"`
}

type SubmissionSummary struct {
    ID string `json:"id"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    GradingStartTime time.Time `json:"grading_start_time"`
}

func (this SubmissionSummary) String() string {
    return util.BaseString(this);
}

func (this *SubmissionSummary) GetScoringInfo() *ScoringInfo {
    return &ScoringInfo{
        ID: this.ID,
        SubmissionTime: this.GradingStartTime,
        RawScore: this.Score,
    };
}

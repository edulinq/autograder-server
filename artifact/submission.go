package artifact

import (
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

// TEST - We don't really need summaries written to disk, just pass them around.
// We may not need them at all.

type TestSubmission struct {
    IgnoreMessages bool `json:"ignore_messages"`
    Result GradedAssignment `json:"result"`
}

type SubmissionSummary struct {
    ID string `json:"id"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    GradingStartTime common.Timestamp `json:"grading_start_time"`
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

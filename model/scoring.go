package model

import (
    "github.com/eriq-augustine/autograder/common"
)

const SCORING_INFO_STRUCT_VERSION = "1.0.0"

type ScoringInfo struct {
    ID string `json:"id"`
    SubmissionTime common.Timestamp `json:"submission-time"`
    UploadTime common.Timestamp `json:"upload-time"`
    RawScore float64 `json:"raw-score"`
    Score float64 `json:"score"`
    Lock bool `json:"lock"`
    LateDayUsage int `json:"late-date-usage"`
    NumDaysLate int `json:"num-days-late"`
    Reject bool `json:"reject"`

    // A distinct key so we can recognize this as an autograder object.
    AutograderStructVersion string `json:"__autograder__version__"`

    // If this object was serialized from an LMS comment, keep the ID.
    LMSCommentID string `json:"-"`
    LMSCommentAuthorID string `json:"-"`
}

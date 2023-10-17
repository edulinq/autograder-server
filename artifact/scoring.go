package artifact

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
    Reject bool `json:"reject"`

    // A distinct key so we can recognize this as an autograder object.
    Autograder int `json:"__autograder__v01__"`
    // If this object was serialized from an LMS comment, keep the ID.
    LMSCommentID string `json:"-"`
    LMSCommentAuthorID string `json:"-"`
}

package artifact

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

// TODO(eriq): Include output.
type GradingResult struct {
    Result *GradedAssignment `json:"result"`
    InputFilesGZip map[string][]byte `json:"input-files-gzip"`
}

// TEST - Rename to SubmissionResult? Get a name the complements the above struct (GradingResult).
type GradedAssignment struct {
    // Information set by the autograder.
    ID string `json:"id"`
    ShortID string `json:"short-id"`
    CourseID string `json:"course-id"`
    AssignmentID string `json:"assignment-id"`
    User string `json:"user"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`

    // Information generally filled out by the grader.
    Name string `json:"name"`
    Questions []GradedQuestion `json:"questions"`
    GradingStartTime common.Timestamp `json:"grading_start_time"`
    GradingEndTime common.Timestamp `json:"grading_end_time"`

    // Additional pass-through information that the grader can use.
    AdditionalInfo map[string]any `json:"additional-info"`
}

type SubmissionHistoryItem struct {
    ID string `json:"id"`
    ShortID string `json:"short-id"`
    CourseID string `json:"course-id"`
    AssignmentID string `json:"assignment-id"`
    User string `json:"user"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    GradingStartTime common.Timestamp `json:"grading_start_time"`
}

type GradedQuestion struct {
    Name string `json:"name"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    Message string `json:"message"`
    GradingStartTime common.Timestamp `json:"grading_start_time"`
    GradingEndTime common.Timestamp `json:"grading_end_time"`
}

func (this GradedAssignment) ToHistoryItem() *SubmissionHistoryItem {
    return &SubmissionHistoryItem{
        ID: this.ID,
        ShortID: this.ShortID,
        CourseID: this.CourseID,
        AssignmentID: this.AssignmentID,
        User: this.User,
        Message: this.Message,
        MaxPoints: this.MaxPoints,
        Score: this.Score,
        GradingStartTime: this.GradingStartTime,
    };
}

func (this GradedAssignment) ToScoringInfo() *ScoringInfo {
    return &ScoringInfo{
        ID: this.ID,
        SubmissionTime: this.GradingStartTime,
        RawScore: this.Score,
    };
}

func (this GradedAssignment) String() string {
    return util.BaseString(this);
}

func (this GradedAssignment) Equals(other GradedAssignment, checkMessages bool) bool {
    if (this.Name != other.Name) {
        return false;
    }

    if ((this.Questions == nil) && (other.Questions == nil)) {
        return true;
    }

    if ((this.Questions == nil) || (other.Questions == nil)) {
        return false;
    }

    if (len(this.Questions) != len(other.Questions)) {
        return false;
    }

    for i := 0; i < len(this.Questions); i++ {
        if (!this.Questions[i].Equals(other.Questions[i], checkMessages)) {
            return false;
        }
    }

    return true;
}

func (this GradedAssignment) Report() string {
    var builder strings.Builder;

    builder.WriteString(fmt.Sprintf("Autograder transcript for assignment: %s.\n", this.Name));
    builder.WriteString(fmt.Sprintf("Grading started at %s and ended at %s.\n", this.GradingStartTime, this.GradingEndTime));

    totalScore := 0.0;
    maxScore := 0.0;

    for _, question := range this.Questions {
        totalScore += question.Score;
        maxScore += question.MaxPoints;

        builder.WriteString(fmt.Sprintf("%s", question.Report()));
    }

    builder.WriteString("\n");
    builder.WriteString(fmt.Sprintf("Total: %s / %s", util.FloatToStr(totalScore), util.FloatToStr(maxScore)));

    return builder.String();
}

// Fill in the MaxPoints and Score fields.
func (this GradedAssignment) ComputePoints() {
    for _, question := range this.Questions {
        this.Score += question.Score;
        this.MaxPoints += question.MaxPoints;
    }
}

// TEST
func (this GradedAssignment) GetSummary(id string, message string) *SubmissionSummary {
    summary := SubmissionSummary{
        ID: id,
        Message: message,
        GradingStartTime: this.GradingStartTime,
    };


    for _, question := range this.Questions {
        summary.Score += question.Score;
        summary.MaxPoints += question.MaxPoints;
    }

    return &summary;
}

func (this GradedQuestion) Report() string {
    var builder strings.Builder;

    builder.WriteString(fmt.Sprintf("%s: %s / %s\n", this.Name, util.FloatToStr(this.Score), util.FloatToStr(this.MaxPoints)));

    if (this.Message != "") {
        for _, line := range strings.Split(this.Message, "\n") {
            builder.WriteString(fmt.Sprintf("    %s\n", strings.TrimSpace(line)));
        }
    }

    return builder.String();
}

func (this GradedQuestion) Equals(other GradedQuestion, checkMessages bool) bool {
    if ((this.Name != other.Name) || (this.MaxPoints != other.MaxPoints) || (this.Score != other.Score)) {
        return false;
    }

    if (checkMessages && (this.Message != other.Message)) {
        return false;
    }

    return true;
}

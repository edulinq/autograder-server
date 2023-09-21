package model

import (
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

type SubmissionSummary struct {
    ID string `json:"id"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    GradingStartTime time.Time `json:"grading_start_time"`
}

type GradedAssignment struct {
    Name string `json:"name"`
    Questions []GradedQuestion `json:"questions"`
    GradingStartTime time.Time `json:"grading_start_time"`
    GradingEndTime time.Time `json:"grading_end_time"`
}

type GradedQuestion struct {
    Name string `json:"name"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    Message string `json:"message"`
    GradingStartTime time.Time `json:"grading_start_time"`
    GradingEndTime time.Time `json:"grading_end_time"`
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
    builder.WriteString(fmt.Sprintf("Total: %s / %s", strFloat(totalScore), strFloat(maxScore)));

    return builder.String();
}

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

    builder.WriteString(fmt.Sprintf("%s: %s / %s\n", this.Name, strFloat(this.Score), strFloat(this.MaxPoints)));

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

func strFloat(value float64) string {
    return strconv.FormatFloat(value, 'f', -1, 64);
}

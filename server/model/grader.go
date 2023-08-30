package model

import (
    "fmt"
    "reflect"
    "strings"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

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

func (this *GradedAssignment) String() string {
    return util.BaseString(this);
}

func (this *GradedAssignment) Equals(other *GradedAssignment) bool {
    if (other == nil) {
        return false;
    }

    if (this == other) {
        return true;
    }

    return (this.Name == other.Name) && (reflect.DeepEqual(this.Questions, other.Questions));
}

func (this *GradedAssignment) Report() string {
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
    builder.WriteString(fmt.Sprintf("Total: %f / %f", totalScore, maxScore));

    return builder.String();
}

func (this *GradedQuestion) Report() string {
    var builder strings.Builder;

    builder.WriteString(fmt.Sprintf("%s: %f / %f\n", this.Name, this.Score, this.MaxPoints));

    if (this.Message != "") {
        for _, line := range strings.Split(this.Message, "\n") {
            builder.WriteString(fmt.Sprintf("    %s\n", strings.TrimSpace(line)));
        }
    }

    return builder.String();
}

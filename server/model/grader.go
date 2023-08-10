package model

import (
    "reflect"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

type GradingResult struct {
    Name string `json:name`
    Start time.Time `json:start`
    End time.Time `json:end`
    Questions []QuestionResult `json:questions`
}

type QuestionResult struct {
    Name string `json:name`
    MaxPoints float64 `json:max_points`
    Score float64 `json:score`
    Message string `json:message`
}

func (this *GradingResult) String() string {
    return util.BaseString(this);
}

func (this *GradingResult) Equals(other *GradingResult) bool {
    if (other == nil) {
        return false;
    }

    if (this == other) {
        return true;
    }

    return (this.Name == other.Name) && (reflect.DeepEqual(this.Questions, other.Questions));
}

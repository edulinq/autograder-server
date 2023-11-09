package db

import (
    "fmt"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/model"
)

func SaveSubmissions(rawCourse model.Course, submissions []*artifact.GradingResult) error {
    course, ok := rawCourse.(*types.Course);
    if (!ok) {
        return fmt.Errorf("Course '%v' is not a db course.", rawCourse);
    }

    return backend.SaveSubmissions(course, submissions);
}

func SaveSubmission(rawCourse model.Course, submission *artifact.GradingResult) error {
    return SaveSubmissions(rawCourse, []*artifact.GradingResult{submission});
}

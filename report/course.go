package report

import (
    "github.com/eriq-augustine/autograder/model2"
)

type CourseScoringReport struct {
    CourseName string `json:"course-name"`
    Assignments []*AssignmentScoringReport `json:"assignments"`
}

func GetCourseScoringReport(course model2.Course) (*CourseScoringReport, error) {
    assignmentReports := make([]*AssignmentScoringReport, 0);

    for _, assignment := range course.GetSortedAssignments() {
        assignmentReport, err := GetAssignmentScoringReport(assignment);
        if (err != nil) {
            return nil, err;
        }

        assignmentReports = append(assignmentReports, assignmentReport);
    }

    report := CourseScoringReport {
        CourseName: course.GetName(),
        Assignments: assignmentReports,
    };

    return &report, nil;
}

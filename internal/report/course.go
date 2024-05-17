package report

import (
    "github.com/edulinq/autograder/internal/model"
)

type CourseScoringReport struct {
    CourseName string `json:"course-name"`
    Assignments []*AssignmentScoringReport `json:"assignments"`
}

func GetCourseScoringReport(course *model.Course) (*CourseScoringReport, error) {
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

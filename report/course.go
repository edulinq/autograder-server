package report

import (
)

type ReportingSources interface {
    GetName() string
    GetReportingSources() []ReportingSource
}

type CourseScoringReport struct {
    CourseName string `json:"course-name"`
    Assignments []*AssignmentScoringReport `json:"assignments"`
}

func GetCourseScoringReport(sources ReportingSources) (*CourseScoringReport, error) {
    assignmentReports := make([]*AssignmentScoringReport, 0);

    for _, source := range sources.GetReportingSources() {
        assignmentReport, err := GetAssignmentScoringReport(source);
        if (err != nil) {
            return nil, err;
        }

        assignmentReports = append(assignmentReports, assignmentReport);
    }

    report := CourseScoringReport {
        CourseName: sources.GetName(),
        Assignments: assignmentReports,
    };

    return &report, nil;
}

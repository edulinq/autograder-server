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

    reportingSources := sources.GetReportingSources();
    for i := (len(reportingSources) - 1); i >= 0; i-- {
        assignmentReport, err := GetAssignmentScoringReport(reportingSources[i]);
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

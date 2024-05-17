package report

import (
    "reflect"
    "testing"

    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/util"
)

func TestCourseReportBase(test *testing.T) {
    course := db.MustGetTestCourse();

    report, err := GetCourseScoringReport(course);
    if (err != nil) {
        test.Fatalf("Failed to get course report: '%v'.", err);
    }

    if (!reflect.DeepEqual(expected, report)) {
        test.Fatalf("Report not as expected.\n--- Expected ---\n%s\n--- Actual ---\n%s\n",
                util.MustToJSONIndent(expected), util.MustToJSONIndent(report));
    }
}

func TestCourseReportHTML(test *testing.T) {
    course := db.MustGetTestCourse();

    report, err := GetCourseScoringReport(course);
    if (err != nil) {
        test.Fatalf("Failed to get course report: '%v'.", err);
    }

    reportHTML, err := report.ToHTML();
    if (err != nil) {
        test.Fatalf("Failed to generate HTML for report: '%v'.", err);
    }

    expectedHTML, err := expected.ToHTML();
    if (err != nil) {
        test.Fatalf("Failed to generate HTML for expected report: '%v'.", err);
    }

    if (expectedHTML != reportHTML) {
        test.Fatalf("Report HTML not as expected.\n--- Expected ---\n%s\n--- Actual ---\n%s\n",
                expectedHTML, reportHTML);
    }
}

var expected *CourseScoringReport = &CourseScoringReport{
    CourseName: "Course 101",
    Assignments: []*AssignmentScoringReport{
        &AssignmentScoringReport{
            AssignmentName: "Homework 0",
            NumberOfSubmissions: 1,
            LatestSubmission: common.MustTimestampFromString("2023-10-15T21:44:33Z"),
            Questions: []*ScoringReportQuestionStats{
                &ScoringReportQuestionStats{
                    QuestionName: "Q1",
                    Min: 1,
                    Max: 1,
                    Median: 1,
                    Mean: 1,
                    StdDev: -1,
                    MinString: "1.00",
                    MaxString: "1.00",
                    MedianString: "1.00",
                    MeanString: "1.00",
                    StdDevString: "NaN",
                },
                &ScoringReportQuestionStats{
                    QuestionName: "Q2",
                    Min: 1,
                    Max: 1,
                    Median: 1,
                    Mean: 1,
                    StdDev: -1,
                    MinString: "1.00",
                    MaxString: "1.00",
                    MedianString: "1.00",
                    MeanString: "1.00",
                    StdDevString: "NaN",
                },
                &ScoringReportQuestionStats{
                    QuestionName: "Style",
                    Min: 0,
                    Max: 0,
                    Median: 0,
                    Mean: 0,
                    StdDev: -1,
                    MinString: "0.00",
                    MaxString: "0.00",
                    MedianString: "0.00",
                    MeanString: "0.00",
                    StdDevString: "NaN",
                },
                &ScoringReportQuestionStats{
                    QuestionName: "<Overall>",
                    Min: 1,
                    Max: 1,
                    Median: 1,
                    Mean: 1,
                    StdDev: -1,
                    MinString: "1.00",
                    MaxString: "1.00",
                    MedianString: "1.00",
                    MeanString: "1.00",
                    StdDevString: "NaN",
                },
            },
        },
    },
};

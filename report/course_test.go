package report

import (
    "reflect"
    "testing"

    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/util"
)

func TestCourseReportBase(test *testing.T) {
    course := db.MustGetTestCourse();

    report, err := GetCourseScoringReport(course);
    if (err != nil) {
        test.Fatalf("Failed to get course report: '%v'.", err);
    }

    if (!reflect.DeepEqual(TestCourseReportExpected, report)) {
        test.Fatalf("Report not as expected.\n--- Expected ---\n%s\n--- Actual ---\n%s\n",
                util.MustToJSONIndent(TestCourseReportExpected), util.MustToJSONIndent(report));
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

    expectedHTML, err := TestCourseReportExpected.ToHTML();
    if (err != nil) {
        test.Fatalf("Failed to generate HTML for expected report: '%v'.", err);
    }

    if (expectedHTML != reportHTML) {
        test.Fatalf("Report HTML not as expected.\n--- Expected ---\n%s\n--- Actual ---\n%s\n",
                expectedHTML, reportHTML);
    }
}

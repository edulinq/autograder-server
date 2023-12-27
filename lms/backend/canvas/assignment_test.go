package canvas

import (
    "testing"
    "time"

    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/util"
)

const TEST_ASSIGNMENT_ID = "98765";

func TestFetchAssignmentBase(test *testing.T) {
    expected := lmstypes.Assignment{
        ID: TEST_ASSIGNMENT_ID,
        Name: "Assignment 0",
        LMSCourseID: "12345",
        DueDate: mustParseTime(test, "2023-10-06T06:59:59Z"),
        MaxPoints: 100.0,
    };

    assignment, err := testBackend.FetchAssignment(TEST_ASSIGNMENT_ID);
    if (err != nil) {
        test.Fatalf("Failed tp fetch assignment: '%v'.", err);
    }

    // Can't compare directly because of time.Time.
    // Use JSON instead.
    expectedJSON := util.MustToJSONIndent(expected);
    actualJSON := util.MustToJSONIndent(assignment);

    if (expectedJSON != actualJSON) {
        test.Fatalf("Assignment not as expected. Expected: '%s', Actual: '%s'.",
                expectedJSON, actualJSON);
    }
}

func mustParseTime(test *testing.T, text string) *time.Time {
    instance, err := time.Parse(time.RFC3339, text);
    if (err != nil) {
        test.Fatalf("Failed to parse time '%s': '%v'.", text, err);
    }

    return &instance;
}

package canvas

import (
    "testing"

    "github.com/edulinq/autograder/internal/lms/lmstypes"
    "github.com/edulinq/autograder/internal/util"
)

var expectedAssignment lmstypes.Assignment = lmstypes.Assignment{
    ID: TEST_ASSIGNMENT_ID,
    Name: "Assignment 0",
    LMSCourseID: "12345",
    DueDate: mustParseTime("2023-10-06T06:59:59Z"),
    MaxPoints: 100.0,
};

func TestFetchAssignmentBase(test *testing.T) {
    assignment, err := testBackend.FetchAssignment(TEST_ASSIGNMENT_ID);
    if (err != nil) {
        test.Fatalf("Failed tp fetch assignment: '%v'.", err);
    }

    // Can't compare directly because of time.Time.
    // Use JSON instead.
    expectedJSON := util.MustToJSONIndent(expectedAssignment);
    actualJSON := util.MustToJSONIndent(assignment);

    if (expectedJSON != actualJSON) {
        test.Fatalf("Assignment not as expected. Expected: '%s', Actual: '%s'.",
                expectedJSON, actualJSON);
    }
}

func TestFetchAssignmentsBase(test *testing.T) {
    assignments, err := testBackend.FetchAssignments();
    if (err != nil) {
        test.Fatalf("Failed tp fetch assignments: '%v'.", err);
    }

    // Can't compare directly because of time.Time.
    // Use JSON instead.
    expectedJSON := util.MustToJSONIndent([]*lmstypes.Assignment{&expectedAssignment});
    actualJSON := util.MustToJSONIndent(assignments);

    if (expectedJSON != actualJSON) {
        test.Fatalf("Assignment not as expected. Expected: '%s', Actual: '%s'.",
                expectedJSON, actualJSON);
    }
}

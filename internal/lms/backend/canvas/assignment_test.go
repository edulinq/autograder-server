package canvas

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var dueDate timestamp.Timestamp = timestamp.MustGuessFromString("2023-10-06T06:59:59Z")
var expectedAssignment lmstypes.Assignment = lmstypes.Assignment{
	ID:          TEST_ASSIGNMENT_ID,
	Name:        "Assignment 0",
	LMSCourseID: "12345",
	DueDate:     &dueDate,
	MaxPoints:   100.0,
}

func TestFetchAssignmentBase(test *testing.T) {
	assignment, err := testBackend.FetchAssignment(TEST_ASSIGNMENT_ID)
	if err != nil {
		test.Fatalf("Failed tp fetch assignment: '%v'.", err)
	}

	if !reflect.DeepEqual(&expectedAssignment, assignment) {
		test.Fatalf("Assignment not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedAssignment), util.MustToJSONIndent(assignment))
	}
}

func TestFetchAssignmentsBase(test *testing.T) {
	assignments, err := testBackend.FetchAssignments()
	if err != nil {
		test.Fatalf("Failed tp fetch assignments: '%v'.", err)
	}

	expected := []*lmstypes.Assignment{&expectedAssignment}

	if !reflect.DeepEqual(expected, assignments) {
		test.Fatalf("Assignment not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(assignments))
	}
}

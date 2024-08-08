package core

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestNewAssignmentInfos(test *testing.T) {
	testCases := []struct {
		course   string
		expected []*AssignmentInfo
	}{
		// Empty
		{"course-without-source", []*AssignmentInfo{}},
		// One Assignment
		{"course101-with-zero-limit", []*AssignmentInfo{
			&AssignmentInfo{"hw0", "Homework 0"},
		}},
		// Multiple Assignments
		{"course-languages", []*AssignmentInfo{
			&AssignmentInfo{"bash", "bash"},
			&AssignmentInfo{"cpp-simple", "cpp-simple"},
			&AssignmentInfo{"java", "java"},
		}},
	}

	for i, testCase := range testCases {
		course := db.MustGetCourse(testCase.course)
		info := NewAssignmentInfos(course.GetSortedAssignments())

		if !reflect.DeepEqual(testCase.expected, info) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(info))
			continue
		}
	}
}

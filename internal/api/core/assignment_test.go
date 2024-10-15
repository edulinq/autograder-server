package core

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestNewAssignmentInfos(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	emptyCourse := &model.Course{ID: "empty-course"}
	db.MustSaveCourse(emptyCourse)

	testCases := []struct {
		course   string
		expected []*AssignmentInfo
	}{
		// Empty
		{"empty-course", []*AssignmentInfo{}},

		// One Assignment
		{"course101", []*AssignmentInfo{
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

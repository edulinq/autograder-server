package db

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestParseServerUserReference(test *testing.T) {
	testCases := []struct {
		reference      string
		output         *UserReference
		errorSubstring string
	}{
		// Email Reference
		{
			"course-student@test.edulinq.org",
			&UserReference{
				Type:  EmailReference,
				Email: "course-student@test.edulinq.org",
			},
			"",
		},
		{
			"-course-student@test.edulinq.org",
			&UserReference{
				Type:    EmailReference,
				Exclude: true,
				Email:   "course-student@test.edulinq.org",
			},
			"",
		},

		// Course Role Reference
		{
			strings.Join([]string{"*", "student"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:           CourseRoleReference,
				Course:         nil,
				CourseUserRole: model.GetCourseUserRole("student"),
			},
			"",
		},
		{
			strings.Join([]string{"-*", "student"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:           CourseRoleReference,
				Exclude:        true,
				Course:         nil,
				CourseUserRole: model.GetCourseUserRole("student"),
			},
			"",
		},
		{
			strings.Join([]string{TEST_COURSE_ID, "student"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:           CourseRoleReference,
				Course:         MustGetTestCourse(),
				CourseUserRole: model.GetCourseUserRole("student"),
			},
			"",
		},
		{
			strings.Join([]string{fmt.Sprintf("-%s", TEST_COURSE_ID), "student"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:           CourseRoleReference,
				Exclude:        true,
				Course:         MustGetTestCourse(),
				CourseUserRole: model.GetCourseUserRole("student"),
			},
			"",
		},

		// Course Reference
		{
			TEST_COURSE_ID,
			&UserReference{
				Type:   CourseReference,
				Course: MustGetTestCourse(),
			},
			"",
		},
		{
			fmt.Sprintf("-%s", TEST_COURSE_ID),
			&UserReference{
				Type:    CourseReference,
				Exclude: true,
				Course:  MustGetTestCourse(),
			},
			"",
		},
		{
			strings.Join([]string{TEST_COURSE_ID, "*"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:   CourseReference,
				Course: MustGetTestCourse(),
			},
			"",
		},
		{
			strings.Join([]string{fmt.Sprintf("-%s", TEST_COURSE_ID), "*"}, USER_REFERENCE_DELIM),
			&UserReference{
				Type:    CourseReference,
				Exclude: true,
				Course:  MustGetTestCourse(),
			},
			"",
		},

		// Server Role Reference
		{
			"user",
			&UserReference{
				Type:           ServerRoleReference,
				ServerUserRole: model.GetServerUserRole("user"),
			},
			"",
		},
		{
			"-user",
			&UserReference{
				Type:           ServerRoleReference,
				Exclude:        true,
				ServerUserRole: model.GetServerUserRole("user"),
			},
			"",
		},

		// All User Reference
		{
			"*",
			&UserReference{
				Type: AllUserReference,
			},
			"",
		},
		{
			"-*",
			&UserReference{
				Type:    AllUserReference,
				Exclude: true,
			},
			"",
		},

		// Errors

		// Accessing Root
		{
			"root",
			nil,
			"User reference cannot target the root user",
		},
		{
			"-root",
			nil,
			"User reference cannot target the root user",
		},
	}

	for i, testCase := range testCases {
		result, err := ParseUserReference(testCase.reference)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%s'.",
						i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.", i, testCase.reference, err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error for reference '%s'.", i, testCase.reference)
			continue
		}

		// Check and clear course information.
		if testCase.output.Course != nil && result.Course != nil {
			if testCase.output.Course.GetID() != result.Course.GetID() {
				test.Errorf("Case %d: Unexpected course ID. Expected: '%s', actual: '%s'.",
					i, testCase.output.Course.GetID(), result.Course.GetID())
				continue
			}

			testCase.output.Course = nil
			result.Course = nil
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

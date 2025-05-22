package db

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestParseCourseUserReference(test *testing.T) {
	testCases := []struct {
		input      []model.CourseUserReferenceInput
		output     *model.CourseUserReference
		workErrors map[string]error
	}{
		// Target Emails
		{
			[]model.CourseUserReferenceInput{"course-student@test.edulinq.org"},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{"-course-student@test.edulinq.org"},
			&model.CourseUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			nil,
		},

		// Target Roles
		{
			[]model.CourseUserReferenceInput{"admin"},
			&model.CourseUserReference{
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{"-admin"},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},

		// All Users
		{
			[]model.CourseUserReferenceInput{"*"},
			&model.CourseUserReference{
				CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{"-*"},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			nil,
		},

		// Complex, Normalization
		{
			[]model.CourseUserReferenceInput{
				"course-student@test.edulinq.org",
				"COURSE-student@test.EDULINQ.org",
				"admin",
				"aDmIn",
			},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{
				"course-student@test.edulinq.org    	",
				"    	course-student@test.edulinq.org",
				"   -admin",
				"-admin	",
			},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},

		// Complex, Non-Overlapping
		{
			[]model.CourseUserReferenceInput{
				"course-student@test.edulinq.org",
				"-course-admin@test.edulinq.org",
				"admin",
				"-owner",
			},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"owner": nil,
				},
			},
			nil,
		},

		// Complex, Overlapping
		{
			[]model.CourseUserReferenceInput{
				"course-student@test.edulinq.org",
				"-course-student@test.edulinq.org",
				"admin",
				"-admin",
			},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				CourseUserRoles: map[string]any{
					"admin": nil,
				},
				ExcludeCourseUserRoles: map[string]any{
					"admin": nil,
				},
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{
				"grader",
				"*",
			},
			&model.CourseUserReference{
				CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			nil,
		},
		{
			[]model.CourseUserReferenceInput{
				"-*",
				"-student",
			},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			nil,
		},

		// Errors

		// Unknown Course User
		{
			[]model.CourseUserReferenceInput{"zzz@test.edulinq.org"},
			nil,
			map[string]error{
				"zzz@test.edulinq.org": fmt.Errorf("Unknown course user: 'zzz@test.edulinq.org'."),
			},
		},
		{
			[]model.CourseUserReferenceInput{"server-user@test.edulinq.org"},
			nil,
			map[string]error{
				"server-user@test.edulinq.org": fmt.Errorf("Unknown course user: 'server-user@test.edulinq.org'."),
			},
		},

		// Unknown Course Role
		{
			[]model.CourseUserReferenceInput{"ZZZ"},
			nil,
			map[string]error{
				"ZZZ": fmt.Errorf("Unknown course role: 'zzz'."),
			},
		},
	}

	testCourse := MustGetTestCourse()

	for i, testCase := range testCases {
		result, workErrors, err := ParseCourseUserReference(testCourse, testCase.input)
		if err != nil {
			test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.",
				i, util.MustToJSONIndent(testCase.output), err.Error())
		}

		if len(testCase.workErrors) != 0 {
			if !reflect.DeepEqual(testCase.workErrors, workErrors) {
				test.Errorf("Case %d: Unexpected work errors. Expected: '%s', Actual: '%s'.",
					i, util.MustToJSONIndent(testCase.workErrors), util.MustToJSONIndent(workErrors))
			}

			continue
		}

		testCase.output.Course = testCourse
		setCourseUserReferenceDefaults(testCase.output)

		// Check and clear course information to pass equality check.
		failed := checkAndClearCourse(test, i, testCase.output, result)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

func setCourseUserReferenceDefaults(reference *model.CourseUserReference) {
	if reference == nil {
		return
	}

	if reference.Emails == nil {
		reference.Emails = make(map[string]any, 0)
	}

	if reference.ExcludeEmails == nil {
		reference.ExcludeEmails = make(map[string]any, 0)
	}

	if reference.CourseUserRoles == nil {
		reference.CourseUserRoles = make(map[string]any, 0)
	}

	if reference.ExcludeCourseUserRoles == nil {
		reference.ExcludeCourseUserRoles = make(map[string]any, 0)
	}
}

func checkAndClearCourse(test *testing.T, i int, expected *model.CourseUserReference, actual *model.CourseUserReference) bool {
	if expected == nil && actual == nil {
		return false
	}

	if expected == nil {
		test.Errorf("Case %d: Unexpected course information. Expected: '%s', Actual: '%s'.",
			i, util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
		return true
	}

	if actual == nil {
		test.Errorf("Case %d: Unexpected course information. Expected: '%s', Actual: '%s'.",
			i, util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
		return true
	}

	if expected.Course != nil && actual.Course != nil {
		if expected.Course.GetID() != actual.Course.GetID() {
			test.Errorf("Case %d: Unexpected course ID. Expected: '%s', actual: '%s'.",
				i, expected.Course.GetID(), actual.Course.GetID())
			return true
		}

		expected.Course = nil
		actual.Course = nil
	}

	return false
}

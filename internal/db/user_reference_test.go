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
		input          []model.ServerUserReferenceInput
		output         *model.ServerUserReference
		errorSubstring string
	}{
		// All Users
		{
			[]model.ServerUserReferenceInput{"*"},
			&model.ServerUserReference{
				ServerUserRoles: model.GetCommonServerUserRoleStrings(),
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{"-*"},
			&model.ServerUserReference{
				ExcludeServerUserRoles: model.GetCommonServerUserRoleStrings(),
			},
			"",
		},

		// Target Email
		{
			[]model.ServerUserReferenceInput{"course-student@test.edulinq.org"},
			&model.ServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{"-course-student@test.edulinq.org"},
			&model.ServerUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},

		// Target Server Role
		{
			[]model.ServerUserReferenceInput{"user"},
			&model.ServerUserReference{
				ServerUserRoles: map[string]model.ServerUserRole{
					"user": model.GetServerUserRole("user"),
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{"-user"},
			&model.ServerUserReference{
				ExcludeServerUserRoles: map[string]model.ServerUserRole{
					"user": model.GetServerUserRole("user"),
				},
			},
			"",
		},

		// All Courses, All Course Roles
		{
			[]model.ServerUserReferenceInput{"*::*"},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:          MustGetTestCourse(),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
					"course-languages": &model.CourseUserReference{
						Course:          MustGetCourse("course-languages"),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{"-*::*"},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:                 MustGetTestCourse(),
						ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
					"course-languages": &model.CourseUserReference{
						Course:                 MustGetCourse("course-languages"),
						ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},

		// All Courses, Target Course Role
		{
			[]model.ServerUserReferenceInput{"*::student"},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
					"course-languages": &model.CourseUserReference{
						Course: MustGetCourse("course-languages"),
						CourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{"-*::student"},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						ExcludeCourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
					"course-languages": &model.CourseUserReference{
						Course: MustGetCourse("course-languages"),
						ExcludeCourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
				},
			},
			"",
		},

		// Target Course, All Course Roles
		{
			[]model.ServerUserReferenceInput{
				model.ServerUserReferenceInput(fmt.Sprintf("%s::*", TEST_COURSE_ID)),
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:          MustGetTestCourse(),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				model.ServerUserReferenceInput(fmt.Sprintf("-%s::*", TEST_COURSE_ID)),
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:                 MustGetTestCourse(),
						ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},

		// Target Course, Target Course Role
		{
			[]model.ServerUserReferenceInput{
				model.ServerUserReferenceInput(fmt.Sprintf("%s::student", TEST_COURSE_ID)),
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				model.ServerUserReferenceInput(fmt.Sprintf("-%s::student", TEST_COURSE_ID)),
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						ExcludeCourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
				},
			},
			"",
		},

		// Complex, Normalization
		{
			[]model.ServerUserReferenceInput{
				"course-student@test.edulinq.org",
				"COURSE-student@test.EDULINQ.org",
				"admin",
				"aDmIn",
				"COURSE101::grader",
				"course101::GRADER",
			},
			&model.ServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[string]model.ServerUserRole{
					"admin": model.GetServerUserRole("admin"),
				},
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				"course-student@test.edulinq.org    	",
				"    	course-student@test.edulinq.org",
				"   admin",
				"admin	",
				"	course101     ::   grader	",
				" course101	::	grader     ",
			},
			&model.ServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[string]model.ServerUserRole{
					"admin": model.GetServerUserRole("admin"),
				},
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
					},
				},
			},
			"",
		},

		// Complex, Non-Overlapping
		{
			[]model.ServerUserReferenceInput{
				"course-student@test.edulinq.org",
				"-course-admin@test.edulinq.org",
				"admin",
				"-owner",
				"course101::grader",
				"-course101::student",
			},
			&model.ServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-admin@test.edulinq.org": nil,
				},
				ServerUserRoles: map[string]model.ServerUserRole{
					"admin": model.GetServerUserRole("admin"),
				},
				ExcludeServerUserRoles: map[string]model.ServerUserRole{
					"owner": model.GetServerUserRole("owner"),
				},
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
						ExcludeCourseUserRoles: map[string]model.CourseUserRole{
							"student": model.GetCourseUserRole("student"),
						},
					},
				},
			},
			"",
		},

		// Complex, Overlapping
		{
			[]model.ServerUserReferenceInput{
				"course-student@test.edulinq.org",
				"-course-student@test.edulinq.org",
				"admin",
				"-admin",
				"course101::grader",
				"-course101::grader",
			},
			&model.ServerUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
				ServerUserRoles: map[string]model.ServerUserRole{
					"admin": model.GetServerUserRole("admin"),
				},
				ExcludeServerUserRoles: map[string]model.ServerUserRole{
					"admin": model.GetServerUserRole("admin"),
				},
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
						ExcludeCourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				"course101::grader",
				"*::grader",
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course: MustGetTestCourse(),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
					},
					"course-languages": &model.CourseUserReference{
						Course: MustGetCourse("course-languages"),
						CourseUserRoles: map[string]model.CourseUserRole{
							"grader": model.GetCourseUserRole("grader"),
						},
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				"course101::grader",
				"course101::*",
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:          MustGetTestCourse(),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				"course101::grader",
				"course101::*",
				"*::grader",
				"*::*",
			},
			&model.ServerUserReference{
				CourseUserReferences: map[string]*model.CourseUserReference{
					TEST_COURSE_ID: &model.CourseUserReference{
						Course:          MustGetTestCourse(),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
					"course-languages": &model.CourseUserReference{
						Course:          MustGetCourse("course-languages"),
						CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
					},
				},
			},
			"",
		},
		{
			[]model.ServerUserReferenceInput{
				"admin",
				"*",
			},
			&model.ServerUserReference{
				ServerUserRoles: model.GetCommonServerUserRoleStrings(),
			},
			"",
		},

		// Errors

		// Accessing Root
		{
			[]model.ServerUserReferenceInput{"root"},
			nil,
			"Unknown server user role 'root'.",
		},
		{
			[]model.ServerUserReferenceInput{"-root"},
			nil,
			"Unknown server user role 'root'.",
		},

		// Unknown Server Role
		{
			[]model.ServerUserReferenceInput{"ZZZ"},
			nil,
			"Unknown server user role 'zzz'.",
		},

		// Unknown Course
		{
			[]model.ServerUserReferenceInput{"ZZZ::*"},
			nil,
			"Unknown course 'zzz'.",
		},

		// Unknown Course Role
		{
			[]model.ServerUserReferenceInput{"*::ZZZ"},
			nil,
			"Unknown course user role 'zzz'.",
		},

		// Invalid Format
		{
			[]model.ServerUserReferenceInput{"foo::bar::baz"},
			nil,
			"Invalid user reference format",
		},
	}

	for i, testCase := range testCases {
		result, err := ParseServerUserReference(testCase.input)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%s'.",
						i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.",
					i, util.MustToJSONIndent(testCase.output), err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error for input '%s'.",
				i, util.MustToJSONIndent(testCase.input))
			continue
		}

		setServerUserReferenceDefaults(testCase.output)

		// Check and clear course information from CourseUserReference map to pass equality check.
		for courseID, courseUserReference := range testCase.output.CourseUserReferences {
			actual, _ := result.CourseUserReferences[courseID]
			failed := checkAndClearCourse(test, i, courseUserReference, actual)
			if failed {
				continue
			}
		}

		if !reflect.DeepEqual(testCase.output, result) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.output), util.MustToJSONIndent(result))
			continue
		}
	}
}

func (this *DBTests) DBTestParseCourseUserReference(test *testing.T) {
	testCases := []struct {
		input          []model.CourseUserReferenceInput
		output         *model.CourseUserReference
		errorSubstring string
	}{
		// Target Emails
		{
			[]model.CourseUserReferenceInput{"course-student@test.edulinq.org"},
			&model.CourseUserReference{
				Emails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},
		{
			[]model.CourseUserReferenceInput{"-course-student@test.edulinq.org"},
			&model.CourseUserReference{
				ExcludeEmails: map[string]any{
					"course-student@test.edulinq.org": nil,
				},
			},
			"",
		},

		// Target Roles
		{
			[]model.CourseUserReferenceInput{"admin"},
			&model.CourseUserReference{
				CourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
			},
			"",
		},
		{
			[]model.CourseUserReferenceInput{"-admin"},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
			},
			"",
		},

		// All Users
		{
			[]model.CourseUserReferenceInput{"*"},
			&model.CourseUserReference{
				CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			"",
		},
		{
			[]model.CourseUserReferenceInput{"-*"},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			"",
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
				CourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
			},
			"",
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
				ExcludeCourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
			},
			"",
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
				CourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
				ExcludeCourseUserRoles: map[string]model.CourseUserRole{
					"owner": model.GetCourseUserRole("owner"),
				},
			},
			"",
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
				CourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
				ExcludeCourseUserRoles: map[string]model.CourseUserRole{
					"admin": model.GetCourseUserRole("admin"),
				},
			},
			"",
		},
		{
			[]model.CourseUserReferenceInput{
				"grader",
				"*",
			},
			&model.CourseUserReference{
				CourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			"",
		},
		{
			[]model.CourseUserReferenceInput{
				"-*",
				"-student",
			},
			&model.CourseUserReference{
				ExcludeCourseUserRoles: model.GetCommonCourseUserRoleStrings(),
			},
			"",
		},

		// Errors

		// Unknown Course User
		{
			[]model.CourseUserReferenceInput{"zzz@test.edulinq.org"},
			nil,
			"Unknown course user 'zzz@test.edulinq.org'.",
		},
		{
			[]model.CourseUserReferenceInput{"server-user@test.edulinq.org"},
			nil,
			"Unknown course user 'server-user@test.edulinq.org'.",
		},

		// Unknown Course Role
		{
			[]model.CourseUserReferenceInput{"ZZZ"},
			nil,
			"Unknown course user role 'zzz'.",
		},
	}

	testCourse := MustGetTestCourse()

	for i, testCase := range testCases {
		result, err := ParseCourseUserReference(testCourse, testCase.input)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%s'.",
						i, testCase.errorSubstring, err.Error())
				}
			} else {
				test.Errorf("Case %d: Failed to parse user reference '%s': '%v'.",
					i, util.MustToJSONIndent(testCase.output), err.Error())
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error for input '%s'.",
				i, util.MustToJSONIndent(testCase.input))
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

func setServerUserReferenceDefaults(reference *model.ServerUserReference) {
	if reference == nil {
		return
	}

	if reference.Emails == nil {
		reference.Emails = make(map[string]any, 0)
	}

	if reference.ExcludeEmails == nil {
		reference.ExcludeEmails = make(map[string]any, 0)
	}

	if reference.ServerUserRoles == nil {
		reference.ServerUserRoles = make(map[string]model.ServerUserRole, 0)
	}

	if reference.ExcludeServerUserRoles == nil {
		reference.ExcludeServerUserRoles = make(map[string]model.ServerUserRole, 0)
	}

	if reference.CourseUserReferences == nil {
		reference.CourseUserReferences = make(map[string]*model.CourseUserReference, 0)
	} else {
		for _, courseUserReference := range reference.CourseUserReferences {
			setCourseUserReferenceDefaults(courseUserReference)
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
		reference.CourseUserRoles = make(map[string]model.CourseUserRole, 0)
	}

	if reference.ExcludeCourseUserRoles == nil {
		reference.ExcludeCourseUserRoles = make(map[string]model.CourseUserRole, 0)
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

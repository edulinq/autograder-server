package courses

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUpsertFromFileSpec(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := filepath.Join(config.GetCourseImportDir(), "testdata")

	emptyDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-empty-")

	missingDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-missing-")
	err := util.RemoveDirent(missingDir)
	if err != nil {
		test.Fatalf("Failed to remove dir: '%v'.", err)
	}

	badJSONDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-badJSON-")
	err = util.WriteFile("{", filepath.Join(badJSONDir, model.COURSE_CONFIG_FILENAME))
	if err != nil {
		test.Fatalf("Failed to write bad JSON course config: '%v'.", err)
	}

	invalidConfigDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-invalidConfig-")
	err = util.WriteFile(`{"id": "_i!@#"}`, filepath.Join(invalidConfigDir, model.COURSE_CONFIG_FILENAME))
	if err != nil {
		test.Fatalf("Failed to write invalid config course config: '%v'.", err)
	}

	// When comparing course-specific results,
	// we will only check if matching courses contains the provided error string.
	// Then, both messages will be zeroed.

	testCases := []struct {
		path                string
		options             CourseUpsertOptions
		clearCourses        bool
		expectedResults     []CourseUpsertResult
		expectedUserError   string
		expectedSystemError bool
	}{
		{
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course101", Success: true, Updated: true},
			},
			"",
			false,
		},
		{
			testdataDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course-languages", Success: true, Updated: true},
				CourseUpsertResult{CourseID: "course-with-lms", Success: true, Updated: true},
				CourseUpsertResult{CourseID: "course-without-source", Success: true, Updated: true},
				CourseUpsertResult{CourseID: "course101", Success: true, Updated: true},
				CourseUpsertResult{CourseID: "course101-with-zero-limit", Success: true, Updated: true},
			},
			"",
			false,
		},
		{
			emptyDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{},
			"",
			false,
		},
		{
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course101", Success: true, Created: true},
			},
			"",
			false,
		},

		// Skips
		{
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				SkipUpdates: true,
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course101", Success: true, Skipped: true},
			},
			"",
			false,
		},
		{
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				SkipCreates: true,
			},
			true,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course101", Success: true, Skipped: true},
			},
			"",
			false,
		},

		// System Error
		{
			missingDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{},
			"",
			true,
		},

		// Course-specific Errors
		{
			badJSONDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: UNKNOWN_COURSE_ID,
					Message:  "Could not unmarshal JSON file",
				},
			},
			"",
			false,
		},
		{
			invalidConfigDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: UNKNOWN_COURSE_ID,
					Message:  "Could not validate course config",
				},
			},
			"",
			false,
		},

		// Permissions
		{
			// Server users cannot update a course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-user@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: "course101",
					Message:  "User does not have sufficient course-level permissions to update course.",
				},
			},
			"",
			false,
		},
		{
			// Course graders cannot update their course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-grader@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: "course101",
					Message:  "User does not have sufficient course-level permissions to update course.",
				},
			},
			"",
			false,
		},
		{
			// Course admins can update their course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-admin@test.edulinq.org"),
			},
			false,
			[]CourseUpsertResult{
				CourseUpsertResult{CourseID: "course101", Success: true, Updated: true},
			},
			"",
			false,
		},
		{
			// Server users cannot create a course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-user@test.edulinq.org"),
			},
			true,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: "course101",
					Message:  "User does not have sufficient server-level permissions to create a course.",
				},
			},
			"",
			false,
		},
		{
			// Course graders cannot create a course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-grader@test.edulinq.org"),
			},
			true,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: "course101",
					Message:  "User does not have sufficient server-level permissions to create a course.",
				},
			},
			"",
			false,
		},
		{
			// Course admins can create a course.
			filepath.Join(testdataDir, "course101"),
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-admin@test.edulinq.org"),
			},
			true,
			[]CourseUpsertResult{
				CourseUpsertResult{
					CourseID: "course101",
					Message:  "User does not have sufficient server-level permissions to create a course.",
				},
			},
			"",
			false,
		},

		// Bad filespec
		{
			"",
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Given FileSpec is not valid: 'A path FileSpec cannot have an empty path.'.",
			false,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		if testCase.clearCourses {
			for _, course := range db.MustGetCourses() {
				err := db.ClearCourse(course)
				if err != nil {
					test.Fatalf("Failed to clear course '%s': '%v'.", course.ID, err)
				}
			}
		}

		filespec := &common.FileSpec{
			Type: common.FILESPEC_TYPE_PATH,
			Path: testCase.path,
		}

		actualResults, actualUserError, err := UpsertFromFileSpec(filespec, testCase.options)
		if err != nil {
			if !testCase.expectedSystemError {
				test.Errorf("Case %d: Unexpected error: '%v'.", i, err)
			}

			continue
		}

		if testCase.expectedSystemError {
			test.Errorf("Case %d: Did not get expected system error.", i)
			continue
		}

		if testCase.expectedUserError != actualUserError {
			test.Errorf("Case %d: Unexpected user error. Expected '%s', Actual: '%s'.", i, testCase.expectedUserError, actualUserError)
			continue
		}

		if len(testCase.expectedResults) != len(actualResults) {
			test.Errorf("Case %d: Unexpected number of results. Expected %d, Actual: %d.", i, len(testCase.expectedResults), len(actualResults))
			continue
		}

		foundError := false
		for j, _ := range actualResults {
			if testCase.expectedResults[j].Message == "" {
				continue
			}

			if !strings.Contains(actualResults[j].Message, testCase.expectedResults[j].Message) {
				test.Errorf("Case %d: Message of result %d does not contain the expected substring. Substring: '%s', Full Message: '%s'.", i, j, testCase.expectedResults[j].Message, actualResults[j].Message)
				foundError = true
				break
			}

			actualResults[j].Message = ""
			testCase.expectedResults[j].Message = ""
		}

		if foundError {
			continue
		}

		if !reflect.DeepEqual(testCase.expectedResults, actualResults) {
			test.Errorf("Case %d: Unexpected results. Expected '%s', Actual: '%s'.", i, util.MustToJSONIndent(testCase.expectedResults), util.MustToJSONIndent(actualResults))
			continue
		}
	}
}

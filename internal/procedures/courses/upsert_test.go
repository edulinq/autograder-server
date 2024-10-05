package courses

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUpsertBase(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := config.GetTestdataDir()
	course101Path := filepath.Join(testdataDir, "course101", model.COURSE_CONFIG_FILENAME)

	emptyDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-empty-")
	missingPath := filepath.Join(emptyDir, model.COURSE_CONFIG_FILENAME)

	missingDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-missing-")
	err := util.RemoveDirent(missingDir)
	if err != nil {
		test.Fatalf("Failed to remove dir: '%v'.", err)
	}

	badJSONDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-badJSON-")
	badJSONPath := filepath.Join(badJSONDir, model.COURSE_CONFIG_FILENAME)
	err = util.WriteFile("{", badJSONPath)
	if err != nil {
		test.Fatalf("Failed to write bad JSON course config: '%v'.", err)
	}

	invalidConfigDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-invalidConfig-")
	invalidConfigPath := filepath.Join(invalidConfigDir, model.COURSE_CONFIG_FILENAME)
	err = util.WriteFile(`{"id": "_i!@#"}`, invalidConfigPath)
	if err != nil {
		test.Fatalf("Failed to write invalid config course config: '%v'.", err)
	}

	testCases := []struct {
		path              string
		options           CourseUpsertOptions
		clearCourses      bool
		expectedResult    *CourseUpsertResult
		expectedErrorPart string
	}{
		{
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			&CourseUpsertResult{
				CourseID:              "course101",
				Success:               true,
				Updated:               true,
				LMSSyncResult:         standardLMSSyncResult,
				BuiltAssignmentImages: standardBuildImages,
			},
			"",
		},
		{
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			&CourseUpsertResult{
				CourseID:              "course101",
				Success:               true,
				Created:               true,
				LMSSyncResult:         emptyLMSSyncResult,
				BuiltAssignmentImages: standardBuildImages,
			},
			"",
		},

		// Dry run will say the same thing as non dry runs.
		{
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				DryRun:      true,
			},
			false,
			&CourseUpsertResult{
				CourseID:              "course101",
				Success:               true,
				Updated:               true,
				LMSSyncResult:         standardLMSSyncResult,
				BuiltAssignmentImages: standardDryRunBuildImages,
			},
			"",
		},
		{
			// This test should not strictly fail,
			// but because of the LMS testing infrastructure it will fail.
			course101Path,
			CourseUpsertOptions{
				DryRun:      true,
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			nil,
			"Could not find course 'course101' for LMS adapter.",
		},

		// Errors
		{
			"",
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Course config path does not point to a file: ''.",
		},
		{
			emptyDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Course config path does not point to a file: '",
		},
		{
			missingDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Course config path does not point to a file: '",
		},
		{
			missingPath,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Course config path does not point to a file: '",
		},
		{
			badJSONPath,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Could not unmarshal JSON file",
		},
		{
			invalidConfigPath,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			nil,
			"Could not validate course config",
		},

		// Permissions
		{
			// Server users cannot update a course.
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-user@test.edulinq.org"),
			},
			false,
			nil,
			"User does not have sufficient course-level permissions to update course.",
		},
		{
			// Server users cannot create a course.
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-user@test.edulinq.org"),
			},
			true,
			nil,
			"User does not have sufficient server-level permissions to create a course.",
		},
		{
			// Course graders cannot update their course.
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-grader@test.edulinq.org"),
			},
			false,
			nil,
			"User does not have sufficient course-level permissions to update course.",
		},
		{
			// Course admins can update their course.
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-admin@test.edulinq.org"),
			},
			false,
			&CourseUpsertResult{
				CourseID:              "course101",
				Success:               true,
				Updated:               true,
				LMSSyncResult:         standardLMSSyncResult,
				BuiltAssignmentImages: standardBuildImages,
			},
			"",
		},
		{
			// Course owners cannot create a course.
			course101Path,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("course-owner@test.edulinq.org"),
			},
			true,
			nil,
			"User does not have sufficient server-level permissions to create a course.",
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

		actualResult, _, err := upsertFromConfigPath(testCase.path, testCase.options)
		if err != nil {
			if testCase.expectedErrorPart == "" {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
				continue
			}

			if !strings.Contains(err.Error(), testCase.expectedErrorPart) {
				test.Errorf("Case %d: Error does not contain expected substring. Substring: '%s', Full Error: '%s'.", i, testCase.expectedErrorPart, err.Error())
			}

			continue
		}

		if testCase.expectedErrorPart != "" {
			test.Errorf("Case %d: Did not get expected error.", i)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedResult, actualResult) {
			test.Errorf("Case %d: Unexpected results. Expected '%s', Actual: '%s'.", i, util.MustToJSONIndent(testCase.expectedResult), util.MustToJSONIndent(actualResult))
			continue
		}
	}
}

var standardLMSSyncResult *model.LMSSyncResult = &model.LMSSyncResult{
	UserSync: []*model.UserOpResult{
		&model.UserOpResult{BaseUserOpResult: model.BaseUserOpResult{Email: "course-admin@test.edulinq.org", Skipped: true}},
		&model.UserOpResult{BaseUserOpResult: model.BaseUserOpResult{Email: "course-grader@test.edulinq.org", Skipped: true}},
		&model.UserOpResult{BaseUserOpResult: model.BaseUserOpResult{Email: "course-other@test.edulinq.org", Skipped: true}},
		&model.UserOpResult{BaseUserOpResult: model.BaseUserOpResult{Email: "course-owner@test.edulinq.org", Skipped: true}},
		&model.UserOpResult{BaseUserOpResult: model.BaseUserOpResult{Email: "course-student@test.edulinq.org", Skipped: true}},
	},
	AssignmentSync: model.NewAssignmentSyncResult(),
}

var emptyLMSSyncResult *model.LMSSyncResult = &model.LMSSyncResult{
	UserSync:       []*model.UserOpResult{},
	AssignmentSync: model.NewAssignmentSyncResult(),
}

var standardBuildImages []string = []string{
	"autograder.course101.hw0",
}

var standardDryRunBuildImages []string = []string{
	"autograder.__autograder_dryrun__course101.hw0",
}

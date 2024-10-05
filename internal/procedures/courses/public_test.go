package courses

import (
	"path/filepath"
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

	testdataDir := config.GetTestdataDir()
	course101Dir := filepath.Join(testdataDir, "course101")
	course101ConfigPath := filepath.Join(course101Dir, model.COURSE_CONFIG_FILENAME)

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
		path                 string
		options              CourseUpsertOptions
		clearCourses         bool
		expectedResultCount  int
		expectedSuccessCount int
		expectedErrorPart    string
	}{
		{
			course101Dir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			1, 1,
			"",
		},
		{
			testdataDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			5, 5,
			"",
		},
		{
			emptyDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			0, 0,
			"",
		},
		{
			course101Dir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			1, 1,
			"",
		},
		{
			course101ConfigPath,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			1, 1,
			"",
		},

		{
			course101Dir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				DryRun:      true,
			},
			false,
			1, 1,
			"",
		},

		// Errors
		{
			missingDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			0, 0,
			"Source dirent for copy does not exist",
		},
		{
			badJSONDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			0, 0,
			"Could not unmarshal JSON file",
		},
		{
			invalidConfigDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			0, 0,
			"Could not validate course config",
		},
		{
			"",
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			false,
			0, 0,
			"A path FileSpec cannot have an empty path.",
		},

		// Creates
		{
			course101Dir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			1, 1,
			"",
		},
		{
			testdataDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			5, 5,
			"",
		},
		{
			emptyDir,
			CourseUpsertOptions{
				ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
			},
			true,
			0, 0,
			"",
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

		actualResults, err := UpsertFromFileSpec(filespec, testCase.options)
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

		if testCase.expectedResultCount != len(actualResults) {
			test.Errorf("Case %d: Unexpected number of results. Expected %d, Actual: %d.", i, testCase.expectedResultCount, len(actualResults))
			continue
		}

		actualSuccessCount := 0
		for _, result := range actualResults {
			if result.Success {
				actualSuccessCount++
			}
		}

		if testCase.expectedSuccessCount != actualSuccessCount {
			test.Errorf("Case %d: Unexpected number of successes. Expected %d, Actual: %d.", i, testCase.expectedSuccessCount, actualSuccessCount)
			continue
		}
	}
}

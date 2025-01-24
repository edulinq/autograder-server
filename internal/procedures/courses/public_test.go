package courses

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type basePublicUpsertTestCase struct {
	options              CourseUpsertOptions
	clearCourses         bool
	expectedResultCount  int
	expectedSuccessCount int
	expectedErrorPart    string
}

func TestUpsertFromZipBlob(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := config.GetTestdataDir()
	course101Dir := filepath.Join(testdataDir, "course101")

	emptyDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-zip-empty-")
	defer util.RemoveDirent(emptyDir)

	testCases := []struct {
		path    string
		corrupt bool
		basePublicUpsertTestCase
	}{
		{
			course101Dir,
			false,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				1, 1,
				"",
			},
		},
		{
			testdataDir,
			false,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				2, 2,
				"",
			},
		},
		{
			emptyDir,
			false,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"",
			},
		},

		// Errors
		{
			course101Dir,
			true,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"Failed to unzip payload",
			},
		},
	}

	for i, testCase := range testCases {
		prepUpsertTest(test, testCase.basePublicUpsertTestCase)

		blob, err := util.ZipToBytes(testCase.path, "", true)
		if err != nil {
			test.Errorf("Case %d: Failed to zip payload: '%v'.", i, err)
			continue
		}

		if testCase.corrupt {
			blob = []byte{0, 1, 2, 3, 4, 5, 6, 7}
		}

		actualResults, err := UpsertFromZipBlob(blob, testCase.options)
		processResult(test, testCase.basePublicUpsertTestCase, actualResults, err, fmt.Sprintf("Case %d: ", i))
	}
}

func TestUpsertFromFileSpec(test *testing.T) {
	defer db.ResetForTesting()

	testdataDir := config.GetTestdataDir()
	course101Dir := filepath.Join(testdataDir, "course101")
	course101ConfigPath := filepath.Join(course101Dir, model.COURSE_CONFIG_FILENAME)

	emptyDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-filespec-empty-")
	defer util.RemoveDirent(emptyDir)

	missingDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-filespec-missing-")
	err := util.RemoveDirent(missingDir)
	if err != nil {
		test.Fatalf("Failed to remove dir: '%v'.", err)
	}

	badJSONDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-filespec-badJSON-")
	defer util.RemoveDirent(badJSONDir)
	err = util.WriteFile("{", filepath.Join(badJSONDir, model.COURSE_CONFIG_FILENAME))
	if err != nil {
		test.Fatalf("Failed to write bad JSON course config: '%v'.", err)
	}

	invalidConfigDir := util.MustMkDirTemp("test-internal.procedures.courses.upsert-filespec-invalidConfig-")
	defer util.RemoveDirent(invalidConfigDir)
	err = util.WriteFile(`{"id": "_i!@#"}`, filepath.Join(invalidConfigDir, model.COURSE_CONFIG_FILENAME))
	if err != nil {
		test.Fatalf("Failed to write invalid config course config: '%v'.", err)
	}

	// Zip up a dir and point to the zip as a filespec.
	tempZipBase := util.MustMkDirTemp("test-internal.procedures.courses.upsert-filespec-zip-")
	defer util.RemoveDirent(tempZipBase)
	tempZipPath := filepath.Join(tempZipBase, "test.zip")
	err = util.Zip(course101Dir, tempZipPath, true)
	if err != nil {
		test.Fatalf("Failed to create temp zip: '%v'.", err)
	}

	testCases := []struct {
		path string
		basePublicUpsertTestCase
	}{
		{
			course101Dir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				1, 1,
				"",
			},
		},
		{
			testdataDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				2, 2,
				"",
			},
		},
		{
			emptyDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"",
			},
		},
		{
			course101Dir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				1, 1,
				"",
			},
		},
		{
			course101ConfigPath,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				1, 1,
				"",
			},
		},
		{
			course101Dir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
					CourseUpsertPublicOptions: CourseUpsertPublicOptions{
						DryRun: true,
					},
				},
				false,
				1, 1,
				"",
			},
		},

		// Point to a Zip File,
		{
			tempZipPath,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				1, 1,
				"",
			},
		},

		// Errors
		{
			missingDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"No targets found",
			},
		},
		{
			badJSONDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"Could not unmarshal JSON file",
			},
		},
		{
			invalidConfigDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"Could not validate course config",
			},
		},
		{
			"",
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				false,
				0, 0,
				"A path FileSpec cannot have an empty path.",
			},
		},

		// Creates
		{
			course101Dir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				true,
				1, 1,
				"",
			},
		},
		{
			testdataDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				true,
				2, 2,
				"",
			},
		},
		{
			emptyDir,
			basePublicUpsertTestCase{
				CourseUpsertOptions{
					ContextUser: db.MustGetServerUser("server-creator@test.edulinq.org"),
				},
				true,
				0, 0,
				"",
			},
		},
	}

	for i, testCase := range testCases {
		prepUpsertTest(test, testCase.basePublicUpsertTestCase)

		filespec := &common.FileSpec{
			Type: common.FILESPEC_TYPE_PATH,
			Path: testCase.path,
		}

		actualResults, err := UpsertFromFileSpec(filespec, testCase.options)
		processResult(test, testCase.basePublicUpsertTestCase, actualResults, err, fmt.Sprintf("Case %d: ", i))
	}
}

func prepUpsertTest(test *testing.T, testCase basePublicUpsertTestCase) {
	db.ResetForTesting()

	if testCase.clearCourses {
		for _, course := range db.MustGetCourses() {
			err := db.ClearCourse(course)
			if err != nil {
				test.Fatalf("Failed to clear course '%s': '%v'.", course.ID, err)
			}
		}
	}
}

func processResult(test *testing.T, testCase basePublicUpsertTestCase, actualResults []CourseUpsertResult, err error, prefix string) {
	if err != nil {
		if testCase.expectedErrorPart == "" {
			test.Errorf("%sGot an unexpected error: '%v'.", prefix, err)
			return
		}

		if !strings.Contains(err.Error(), testCase.expectedErrorPart) {
			test.Errorf("%sError does not contain expected substring. Substring: '%s', Full Error: '%s'.", prefix, testCase.expectedErrorPart, err.Error())
		}

		return
	}

	if testCase.expectedErrorPart != "" {
		test.Errorf("%sDid not get expected error.", prefix)
		return
	}

	if testCase.expectedResultCount != len(actualResults) {
		test.Errorf("%sUnexpected number of results. Expected %d, Actual: %d.", prefix, testCase.expectedResultCount, len(actualResults))
		return
	}

	actualSuccessCount := 0
	for _, result := range actualResults {
		if result.Success {
			actualSuccessCount++
		}
	}

	if testCase.expectedSuccessCount != actualSuccessCount {
		test.Errorf("%sUnexpected number of successes. Expected %d, Actual: %d.", prefix, testCase.expectedSuccessCount, actualSuccessCount)
		return
	}
}

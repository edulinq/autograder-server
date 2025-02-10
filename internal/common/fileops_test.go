package common

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestFileOpValidateBase(test *testing.T) {
	// Note that the expected file op will not be validated (it should be constructed valid).
	testCases := []struct {
		operation      *FileOperation
		expected       *FileOperation
		errorSubstring string
	}{
		// Base
		{
			NewFileOperation([]string{"copy", "a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"cp", "a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"move", "a", "b"}),
			NewFileOperation([]string{"move", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"mv", "a", "b"}),
			NewFileOperation([]string{"move", "a", "b"}),
			"",
		},

		// Casing
		{
			NewFileOperation([]string{"Copy", "a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"COPY", "a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"CoPy", "a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},

		// Normalize Paths
		{
			NewFileOperation([]string{"copy", "c/../a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"copy", "a/", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"copy", "a//b", "b"}),
			NewFileOperation([]string{"copy", "a/b", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"copy", "./a", "b"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},

		// Errors

		// Empty
		{
			nil,
			NewFileOperation(nil),
			"File operation is nil",
		},
		{
			NewFileOperation(nil),
			NewFileOperation(nil),
			"File operation is empty",
		},
		{
			NewFileOperation([]string{}),
			NewFileOperation(nil),
			"File operation is empty",
		},

		// Number of Args
		{
			NewFileOperation([]string{"copy", "a"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"copy", "a", "b", "c"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"move", "a"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"move", "a", "b", "c"}),
			nil,
			"Incorrect number of arguments",
		},

		// Unknown Command
		{
			NewFileOperation([]string{"zzz", "a", "b"}),
			nil,
			"Unknown file operation",
		},
		{
			NewFileOperation([]string{"mkdir", "a"}),
			nil,
			"Unknown file operation",
		},

		// Path Errors
		{
			NewFileOperation([]string{"copy", "a\\b", "b"}),
			nil,
			"contains a backslash",
		},
		{
			NewFileOperation([]string{"copy", "/a", "b"}),
			nil,
			"is an absolute path",
		},
		{
			NewFileOperation([]string{"copy", "..", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", "../a", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", "a/../..", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", "a/../../b", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", ".", "b"}),
			nil,
			"cannot point just to the current directory",
		},
		{
			NewFileOperation([]string{"copy", "a/..", "b"}),
			nil,
			"cannot point just to the current directory",
		},
	}

	for i, testCase := range testCases {
		err := testCase.operation.Validate()
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate operation '%+v': '%v'.", i, testCase.operation, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error.", i)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.operation) {
			test.Errorf("Case %d: Operation not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, testCase.operation)
			continue
		}
	}
}

func TestFileOpToUnix(test *testing.T) {
	baseDir := "/tmp/test"

	testCases := []struct {
		Op       FileOperation
		Expected string
	}{
		{FileOperation([]string{"cp", "a", "b"}), "cp -r /tmp/test/a /tmp/test/b"},
		{FileOperation([]string{"mv", "a", "b"}), "mv /tmp/test/a /tmp/test/b"},

		{FileOperation([]string{"CP", "a", "b"}), "cp -r /tmp/test/a /tmp/test/b"},
		{FileOperation([]string{"MV", "a", "b"}), "mv /tmp/test/a /tmp/test/b"},

		{FileOperation([]string{"cp", "a A", "b B"}), "cp -r '/tmp/test/a A' '/tmp/test/b B'"},
		{FileOperation([]string{"mv", "a A", "b B"}), "mv '/tmp/test/a A' '/tmp/test/b B'"},

		{FileOperation([]string{"cp", "\"a\"", "'b'"}), "cp -r '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
		{FileOperation([]string{"mv", "\"a\"", "'b'"}), "mv '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
	}

	for i, testCase := range testCases {
		err := testCase.Op.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate operation '%+v': '%v'.", i, testCase.Op, err)
			continue
		}

		actual := testCase.Op.ToUnix(baseDir)
		if testCase.Expected != actual {
			test.Errorf("Case %d: Unexpected UNIX command. Expected `%s`, Actual: `%s`.", i, testCase.Expected, actual)
			continue
		}
	}
}

func TestFileOpCopyBase(test *testing.T) {
	tempDir := prepTempDir(test)
	defer util.RemoveDirent(tempDir)

	op := FileOperation([]string{"cp", "a.txt", "b.txt"})

	err := op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	// Getting the hash afer the operation will ensure the file still exists.
	expectedHash, err := util.MD5FileHex(filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to get expected hash: '%v'.", err)
	}

	actualHash, err := util.MD5FileHex(filepath.Join(tempDir, "b.txt"))
	if err != nil {
		test.Fatalf("Failed to get actual hash: '%v'.", err)
	}

	if expectedHash != actualHash {
		test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
	}
}

func TestFileOpMoveBase(test *testing.T) {
	tempDir := prepTempDir(test)
	defer util.RemoveDirent(tempDir)

	op := FileOperation([]string{"mv", "a.txt", "b.txt"})

	err := op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	expectedHash, err := util.MD5FileHex(filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to get expected hash: '%v'.", err)
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	if util.PathExists(filepath.Join(tempDir, "a.txt")) {
		test.Fatalf("Source of move still exists.")
	}

	actualHash, err := util.MD5FileHex(filepath.Join(tempDir, "b.txt"))
	if err != nil {
		test.Fatalf("Failed to get actual hash: '%v'.", err)
	}

	if expectedHash != actualHash {
		test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
	}
}

func prepTempDir(test *testing.T) string {
	tempDir, err := util.MkDirTemp("autograder-testing-fileop-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}

	err = util.WriteFile("AAA\n", filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to write test file: '%v'.", err)
	}

	return tempDir
}

package util

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
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
		{
			NewFileOperation([]string{"make-dir", "a"}),
			NewFileOperation([]string{"make-dir", "a"}),
			"",
		},
		{
			NewFileOperation([]string{"mkdir", "a"}),
			NewFileOperation([]string{"make-dir", "a"}),
			"",
		},
		{
			NewFileOperation([]string{"remove", "a"}),
			NewFileOperation([]string{"remove", "a"}),
			"",
		},
		{
			NewFileOperation([]string{"rm", "a"}),
			NewFileOperation([]string{"remove", "a"}),
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
		{
			NewFileOperation([]string{"make-dir"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"make-dir", "a", "b"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"remove"}),
			nil,
			"Incorrect number of arguments",
		},
		{
			NewFileOperation([]string{"remove", "a", "b"}),
			nil,
			"Incorrect number of arguments",
		},

		// Unknown Command
		{
			NewFileOperation([]string{"zzz", "a", "b"}),
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
			test.Errorf("Case %d: Did not get expected error on '%+v'.", i, testCase.operation)
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
		{FileOperation([]string{"mkdir", "a"}), "mkdir -p /tmp/test/a"},
		{FileOperation([]string{"mkdir", "a/b"}), "mkdir -p /tmp/test/a/b"},
		{FileOperation([]string{"rm", "a"}), "rm -rf /tmp/test/a"},
		{FileOperation([]string{"rm", "a/b"}), "rm -rf /tmp/test/a/b"},

		{FileOperation([]string{"cp", "a A", "b B"}), "cp -r '/tmp/test/a A' '/tmp/test/b B'"},
		{FileOperation([]string{"mv", "a A", "b B"}), "mv '/tmp/test/a A' '/tmp/test/b B'"},
		{FileOperation([]string{"mkdir", "a A"}), "mkdir -p '/tmp/test/a A'"},
		{FileOperation([]string{"rm", "a A"}), "rm -rf '/tmp/test/a A'"},

		{FileOperation([]string{"cp", "\"a\"", "'b'"}), "cp -r '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
		{FileOperation([]string{"mv", "\"a\"", "'b'"}), "mv '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
		{FileOperation([]string{"mkdir", "\"a\"/'b'"}), "mkdir -p '/tmp/test/\"a\"/'\"'\"'b'\"'\"''"},
		{FileOperation([]string{"rm", "\"a\"/'b'"}), "rm -rf '/tmp/test/\"a\"/'\"'\"'b'\"'\"''"},
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
	defer RemoveDirent(tempDir)

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
	expectedHash, err := MD5FileHex(filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to get expected hash: '%v'.", err)
	}

	actualHash, err := MD5FileHex(filepath.Join(tempDir, "b.txt"))
	if err != nil {
		test.Fatalf("Failed to get actual hash: '%v'.", err)
	}

	if expectedHash != actualHash {
		test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
	}
}

func TestFileOpCopyGlob(test *testing.T) {
	tempDir := prepTempDir(test)
	defer RemoveDirent(tempDir)

	copySubDir := "cp-sub-dir"
	MustMkDir(filepath.Join(tempDir, copySubDir))

	err := WriteFile("CCC\n", filepath.Join(tempDir, "c.txt"))
	if err != nil {
		test.Fatalf("Failed to write test file: '%v'.", err)
	}

	op := FileOperation([]string{"cp", "*.txt", copySubDir})

	err = op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	for _, filename := range []string{"a.txt", "c.txt"} {
		expectedHash, err := MD5FileHex(filepath.Join(tempDir, filename))
		if err != nil {
			test.Fatalf("Failed to get expected hash: '%v'.", err)
		}

		actualHash, err := MD5FileHex(filepath.Join(tempDir, copySubDir, filename))
		if err != nil {
			test.Fatalf("Failed to get actual hash: '%v'.", err)
		}

		if expectedHash != actualHash {
			test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
		}
	}
}

func TestFileOpMoveBase(test *testing.T) {
	tempDir := prepTempDir(test)
	defer RemoveDirent(tempDir)

	op := FileOperation([]string{"mv", "a.txt", "b.txt"})

	err := op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	expectedHash, err := MD5FileHex(filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to get expected hash: '%v'.", err)
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	if PathExists(filepath.Join(tempDir, "a.txt")) {
		test.Fatalf("Source of move still exists.")
	}

	actualHash, err := MD5FileHex(filepath.Join(tempDir, "b.txt"))
	if err != nil {
		test.Fatalf("Failed to get actual hash: '%v'.", err)
	}

	if expectedHash != actualHash {
		test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
	}
}

func TestFileOpMoveGlobBase(test *testing.T) {
	tempDir := prepTempDir(test)
	defer RemoveDirent(tempDir)

	op := FileOperation([]string{"mv", "a*", "b.txt"})

	err := op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	expectedHash, err := MD5FileHex(filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to get expected hash: '%v'.", err)
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	if PathExists(filepath.Join(tempDir, "a.txt")) {
		test.Fatalf("Source of move still exists.")
	}

	actualHash, err := MD5FileHex(filepath.Join(tempDir, "b.txt"))
	if err != nil {
		test.Fatalf("Failed to get actual hash: '%v'.", err)
	}

	if expectedHash != actualHash {
		test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash)
	}
}

func TestFileOpMoveGlobDir(test *testing.T) {
	tempDir := prepTempDir(test)
	defer RemoveDirent(tempDir)

	moveSubDirname := "mv-sub-dir"
	MustMkDir(filepath.Join(tempDir, moveSubDirname))

	err := WriteFile("CCC\n", filepath.Join(tempDir, "c.txt"))
	if err != nil {
		test.Fatalf("Failed to write test file: '%v'.", err)
	}

	op := FileOperation([]string{"mv", "*.txt", moveSubDirname})

	err = op.Validate()
	if err != nil {
		test.Fatalf("Failed to validate: '%v'.", err)
	}

	filenames := []string{"a.txt", "c.txt"}
	expectedHash := make(map[string]string, 0)
	for _, filename := range filenames {
		expectedHash[filename], err = MD5FileHex(filepath.Join(tempDir, filename))
		if err != nil {
			test.Fatalf("Failed to get expected hash: '%v'.", err)
		}
	}

	err = op.Exec(tempDir)
	if err != nil {
		test.Fatalf("Failed to exec: '%v'.", err)
	}

	for _, filename := range filenames {
		if PathExists(filepath.Join(tempDir, filename)) {
			test.Fatalf("Source of move still exists.")
		}

		actualHash, err := MD5FileHex(filepath.Join(tempDir, moveSubDirname, filename))
		if err != nil {
			test.Fatalf("Failed to get actual hash: '%v'.", err)
		}

		if expectedHash[filename] != actualHash {
			test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash[filename], actualHash)
		}
	}
}

func TestFileOpMkdirBase(test *testing.T) {
	alreadyExistsDirname := "already_exists"
	alreadyExistsFilename := "already_exists.txt"

	testCases := []struct {
		path           string
		errorSubstring string
	}{
		{
			"a",
			"",
		},
		{
			"a/b",
			"",
		},
		{
			"a/../b",
			"",
		},
		{
			alreadyExistsDirname,
			"",
		},
		{
			alreadyExistsDirname + "/a",
			"",
		},
		{
			alreadyExistsFilename,
			"not a directory",
		},
		{
			alreadyExistsFilename + "/a",
			"not a directory",
		},
	}

	for i, testCase := range testCases {
		op := NewFileOperation([]string{"mkdir", testCase.path})
		err := op.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate op '%+v': '%v'.", i, op, err)
			continue
		}

		tempDir := MustMkDirTemp("testing-fileop-mkdir-")
		defer RemoveDirent(tempDir)

		// Make some existing entries.
		MustMkDir(filepath.Join(tempDir, alreadyExistsDirname))
		MustCreateFile(filepath.Join(tempDir, alreadyExistsFilename))

		err = op.Exec(tempDir)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to exec '%+v': '%v'.", i, op, err)
			}

			continue
		}

		expectedPath := filepath.Join(tempDir, testCase.path)
		if !IsDir(expectedPath) {
			test.Errorf("Case %d: Target directory does not exist (or is not a dir) '%s'.", i, expectedPath)
			continue
		}
	}
}

func TestFileOpRemoveBase(test *testing.T) {
	alreadyExistsDirname := "already_exists"
	alreadyExistsFilename := "already_exists.txt"
	alreadyExistsFilePosixRelpath := alreadyExistsDirname + "/" + alreadyExistsFilename
	alreadyExistsFileRelpath := filepath.Join(alreadyExistsDirname, alreadyExistsFilename)

	testCases := []struct {
		path           string
		errorSubstring string
	}{
		{
			"a",
			"",
		},
		{
			"a/b",
			"",
		},
		{
			"a/../b",
			"",
		},
		{
			alreadyExistsDirname,
			"",
		},
		{
			alreadyExistsDirname + "/a",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			"",
		},
	}

	for i, testCase := range testCases {
		op := NewFileOperation([]string{"rm", testCase.path})
		err := op.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate op '%+v': '%v'.", i, op, err)
			continue
		}

		tempDir := MustMkDirTemp("testing-fileop-mkdir-")
		defer RemoveDirent(tempDir)

		// Make some existing entries.
		MustMkDir(filepath.Join(tempDir, alreadyExistsDirname))
		MustCreateFile(filepath.Join(tempDir, alreadyExistsFileRelpath))

		err = op.Exec(tempDir)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to exec '%+v': '%v'.", i, op, err)
			}

			continue
		}

		expectedPath := filepath.Join(tempDir, testCase.path)
		if PathExists(expectedPath) {
			test.Errorf("Case %d: Target path exists when is should have been removed '%s'.", i, expectedPath)
			continue
		}
	}
}

func prepTempDir(test *testing.T) string {
	tempDir, err := MkDirTemp("autograder-testing-fileop-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}

	err = WriteFile("AAA\n", filepath.Join(tempDir, "a.txt"))
	if err != nil {
		test.Fatalf("Failed to write test file: '%v'.", err)
	}

	return tempDir
}

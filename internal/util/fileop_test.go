package util

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var (
	alreadyExistsDirname             = "already_exists"
	alreadyExistsFilename            = "already_exists.txt"
	alreadyExistsFilePosixRelpath    = alreadyExistsDirname + "/" + alreadyExistsFilename
	alreadyExistsFileRelpath         = filepath.Join(alreadyExistsDirname, alreadyExistsFilename)
	alreadyExistsFilenameAlt         = "already_exists_alt.txt"
	alreadyExistsFileAltPosixRelpath = alreadyExistsDirname + "/" + alreadyExistsFilenameAlt
	alreadyExistsFileAltRelpath      = filepath.Join(alreadyExistsDirname, alreadyExistsFilenameAlt)
	startingEmptyDirname             = "empty_start"
)

func TestFileOpValidateBase(test *testing.T) {
	// Note that the expected file operation will not be validated (it should be constructed valid).
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
			NewFileOperation([]string{"copy", " a ", "\tb\n"}),
			NewFileOperation([]string{"copy", "a", "b"}),
			"",
		},
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
		{
			NewFileOperation([]string{"copy", "*/*/..", "b"}),
			NewFileOperation([]string{"copy", "*", "b"}),
			"",
		},

		// Glob Paths
		{
			NewFileOperation([]string{"copy", "a/*", "b"}),
			NewFileOperation([]string{"copy", "a/*", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"copy", "a/?", "b"}),
			NewFileOperation([]string{"copy", "a/?", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"move", "a/*", "b"}),
			NewFileOperation([]string{"move", "a/*", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"move", "a/?", "b"}),
			NewFileOperation([]string{"move", "a/?", "b"}),
			"",
		},
		{
			NewFileOperation([]string{"copy", "a", "*/[a-z"}),
			NewFileOperation([]string{"copy", "a", "*/[a-z"}),
			"",
		},
		{
			NewFileOperation([]string{"make-dir", "*/[a-z"}),
			NewFileOperation([]string{"make-dir", "*/[a-z"}),
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
			NewFileOperation([]string{"copy", "*/../..", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", "*/../../*", "b"}),
			nil,
			"points outside of the its base directory",
		},
		{
			NewFileOperation([]string{"copy", "*/../../*/*", "b"}),
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
		{
			NewFileOperation([]string{"copy", "*/..", "b"}),
			nil,
			"cannot point just to the current directory",
		},
		{
			NewFileOperation([]string{"copy", "*/[a-z", "b"}),
			nil,
			"invalid path pattern",
		},
	}

	for i, testCase := range testCases {
		err := testCase.operation.Validate()
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
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
	testCases := []struct {
		source         string
		dest           string
		errorSubstring string
	}{
		{
			alreadyExistsFilePosixRelpath,
			"a",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			"a.txt",
			"",
		},
		{
			alreadyExistsDirname,
			"a",
			"",
		},
		{
			alreadyExistsDirname,
			"a.txt",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			alreadyExistsDirname + "/a",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			"a/b",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			alreadyExistsDirname,
			"",
		},
		{
			"a",
			"b",
			"Unable to find source path",
		},
		{
			alreadyExistsDirname,
			alreadyExistsFilePosixRelpath,
			"already exists",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"cp", testCase.source, testCase.dest}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			expectedSource := filepath.Join(tempDir, testCase.source)
			expectedDest := filepath.Join(tempDir, testCase.dest)

			if !PathExists(expectedDest) {
				test.Errorf("Case %d: Dest does not exist '%s'.", i, expectedDest)
				return
			}

			if !PathExists(expectedSource) {
				test.Errorf("Case %d: Source does not exist '%s'.", i, expectedSource)
				return
			}
		})
	}
}

func TestFileOpCopyGlob(test *testing.T) {
	testCases := []struct {
		source         string
		dest           string
		expectedExists []string
		errorSubstring string
	}{
		{
			alreadyExistsDirname + "/*",
			startingEmptyDirname,
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join(startingEmptyDirname, alreadyExistsFilename),
				filepath.Join(startingEmptyDirname, alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			startingEmptyDirname,
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join(startingEmptyDirname, alreadyExistsFilename),
				filepath.Join(startingEmptyDirname, alreadyExistsFilenameAlt),
			},
			"",
		},
		// Note, the entire directory is copied into startingEmptyDirname.
		{
			alreadyExistsDirname,
			startingEmptyDirname,
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join(startingEmptyDirname, alreadyExistsFileRelpath),
				filepath.Join(startingEmptyDirname, alreadyExistsFileAltRelpath),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			alreadyExistsDirname,
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			"a",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			"a.txt",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join("a.txt", alreadyExistsFilename),
				filepath.Join("a.txt", alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			"a",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			"a.txt",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join("a.txt", alreadyExistsFilename),
				filepath.Join("a.txt", alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			alreadyExistsDirname + "/*.txt",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join(alreadyExistsDirname, "*.txt", alreadyExistsFilename),
				filepath.Join(alreadyExistsDirname, "*.txt", alreadyExistsFilenameAlt),
			},
			"",
		},
		{
			"*",
			"a",
			[]string{
				alreadyExistsFilename,
				alreadyExistsDirname,
				startingEmptyDirname,
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsDirname),
				filepath.Join("a", startingEmptyDirname),
				filepath.Join("a", alreadyExistsFileRelpath),
				filepath.Join("a", alreadyExistsFileAltRelpath),
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			alreadyExistsFilename,
			[]string{},
			"Failed to create dest dir",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"cp", testCase.source, testCase.dest}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			for _, relExpectedExists := range testCase.expectedExists {
				expectedExists := filepath.Join(tempDir, relExpectedExists)

				if !PathExists(expectedExists) {
					test.Errorf("Case %d: A path does not exist when it should '%s'.", i, expectedExists)
					return
				}
			}
		})
	}
}

func TestFileOpMoveBase(test *testing.T) {
	testCases := []struct {
		source         string
		dest           string
		errorSubstring string
	}{
		{
			alreadyExistsFilePosixRelpath,
			"a",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			"a.txt",
			"",
		},
		{
			alreadyExistsDirname,
			"a",
			"",
		},
		{
			alreadyExistsDirname,
			"a.txt",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			alreadyExistsDirname + "/a",
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			"a/b",
			"",
		},
		{
			"a",
			"b",
			"Unable to find source path",
		},
		// This case outputs a slightly different error on some Mac versions.
		// Let the error substring be very general for this case.
		{
			alreadyExistsDirname,
			alreadyExistsFilePosixRelpath,
			"rename",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"mv", testCase.source, testCase.dest}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			expectedSource := filepath.Join(tempDir, testCase.source)
			expectedDest := filepath.Join(tempDir, testCase.dest)

			if !PathExists(expectedDest) {
				test.Errorf("Case %d: Dest does not exist '%s'.", i, expectedDest)
				return
			}

			if PathExists(expectedSource) {
				test.Errorf("Case %d: Source exists '%s'.", i, expectedSource)
				return
			}
		})
	}
}

func TestFileOpMoveGlob(test *testing.T) {
	testCases := []struct {
		source            string
		dest              string
		expectedExists    []string
		expectedNotExists []string
		errorSubstring    string
	}{
		{
			alreadyExistsDirname + "/*",
			startingEmptyDirname,
			[]string{
				filepath.Join(startingEmptyDirname, alreadyExistsFilename),
				filepath.Join(startingEmptyDirname, alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			startingEmptyDirname,
			[]string{
				filepath.Join(startingEmptyDirname, alreadyExistsFilename),
				filepath.Join(startingEmptyDirname, alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		// Note, the entire directory is copied into startingEmptyDirname.
		{
			alreadyExistsDirname,
			startingEmptyDirname,
			[]string{
				filepath.Join(startingEmptyDirname, alreadyExistsFileRelpath),
				filepath.Join(startingEmptyDirname, alreadyExistsFileAltRelpath),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			alreadyExistsDirname,
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			[]string{},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			"a",
			[]string{
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			"a.txt",
			[]string{
				filepath.Join("a.txt", alreadyExistsFilename),
				filepath.Join("a.txt", alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			"a",
			[]string{
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			"a.txt",
			[]string{
				filepath.Join("a.txt", alreadyExistsFilename),
				filepath.Join("a.txt", alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			alreadyExistsDirname + "/*.txt",
			[]string{
				filepath.Join(alreadyExistsDirname, "*.txt", alreadyExistsFilename),
				filepath.Join(alreadyExistsDirname, "*.txt", alreadyExistsFilenameAlt),
			},
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			"*",
			"a",
			[]string{
				filepath.Join("a", alreadyExistsFilename),
				filepath.Join("a", alreadyExistsDirname),
				filepath.Join("a", startingEmptyDirname),
				filepath.Join("a", alreadyExistsFileRelpath),
				filepath.Join("a", alreadyExistsFileAltRelpath),
			},
			[]string{
				alreadyExistsFilename,
				alreadyExistsDirname,
				startingEmptyDirname,
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			alreadyExistsFilename,
			[]string{},
			[]string{},
			"Failed to create dest dir",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"mv", testCase.source, testCase.dest}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			for _, relExpectedExists := range testCase.expectedExists {
				expectedExists := filepath.Join(tempDir, relExpectedExists)

				if !PathExists(expectedExists) {
					test.Errorf("Case %d: A path does not exist when it should '%s'.", i, expectedExists)
					return
				}
			}

			for _, relExpectedNotExists := range testCase.expectedNotExists {
				expectedNotExists := filepath.Join(tempDir, relExpectedNotExists)

				if PathExists(expectedNotExists) {
					test.Errorf("Case %d: A path exists when it should not '%s'.", i, expectedNotExists)
					return
				}
			}
		})
	}
}

func TestFileOpMkdirBase(test *testing.T) {
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
			"not a directory",
		},
		{
			alreadyExistsFilePosixRelpath + "/a",
			"not a directory",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"mkdir", testCase.path}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			expectedPath := filepath.Join(tempDir, testCase.path)
			if !IsDir(expectedPath) {
				test.Errorf("Case %d: Target directory does not exist (or is not a dir) '%s'.", i, expectedPath)
			}
		})
	}
}

func TestFileOpRemoveBase(test *testing.T) {
	testCases := []struct {
		path           string
		expectedPaths  []string
		errorSubstring string
	}{
		{
			"a",
			[]string{"a"},
			"",
		},
		{
			"a/b",
			[]string{"a/b"},
			"",
		},
		{
			"a/../b",
			[]string{"b"},
			"",
		},
		{
			alreadyExistsDirname,
			[]string{alreadyExistsDirname},
			"",
		},
		{
			alreadyExistsDirname + "/a",
			[]string{alreadyExistsDirname + "/a"},
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			[]string{alreadyExistsFilePosixRelpath},
			"",
		},
		{
			alreadyExistsFilePosixRelpath,
			[]string{alreadyExistsFilePosixRelpath},
			"",
		},
		{
			alreadyExistsDirname + "/*",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			alreadyExistsDirname + "/*.txt",
			[]string{
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
		{
			"*",
			[]string{
				alreadyExistsFilename,
				alreadyExistsDirname,
				startingEmptyDirname,
				alreadyExistsFileRelpath,
				alreadyExistsFileAltRelpath,
			},
			"",
		},
	}

	for i, testCase := range testCases {
		rawOperation := []string{"rm", testCase.path}

		runFileOpExecTest(test, fmt.Sprintf("Case %d", i), rawOperation, testCase.errorSubstring, func(tempDir string) {
			for _, relExpectedPath := range testCase.expectedPaths {
				expectedPath := filepath.Join(tempDir, relExpectedPath)
				if PathExists(expectedPath) {
					test.Errorf("Case %d: Target path exists when is should have been removed '%s'.", i, expectedPath)
				}
			}
		})
	}
}

func runFileOpExecTest(test *testing.T, prefix string, rawOperation []string, errorSubstring string, postExec func(tempDir string)) {
	operation := NewFileOperation(rawOperation)
	err := operation.Validate()
	if err != nil {
		test.Errorf("%s: Failed to validate operation '%+v': '%v'.", prefix, operation, err)
		return
	}

	tempDir := MustMkDirTemp("testing-fileop-exec-")
	defer RemoveDirent(tempDir)

	// Make some existing entries.
	MustMkDir(filepath.Join(tempDir, alreadyExistsDirname))
	MustCreateFile(filepath.Join(tempDir, alreadyExistsFilename))
	MustCreateFile(filepath.Join(tempDir, alreadyExistsFileRelpath))
	MustCreateFile(filepath.Join(tempDir, alreadyExistsFileAltRelpath))
	MustMkDir(filepath.Join(tempDir, startingEmptyDirname))

	err = operation.Exec(tempDir)
	if err != nil {
		if errorSubstring != "" {
			if !strings.Contains(err.Error(), errorSubstring) {
				test.Errorf("%s: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", prefix, errorSubstring, err)
			}
		} else {
			test.Errorf("%s: Failed to exec '%+v': '%v'.", prefix, operation, err)
		}

		return
	}

	if errorSubstring != "" {
		test.Errorf("%s: Did not get expected error '%s'.", prefix, errorSubstring)
		return
	}

	postExec(tempDir)
}

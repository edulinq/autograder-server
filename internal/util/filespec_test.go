package util

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type testCaseParseValidation struct {
	Text             string
	Parses           bool
	Validates        bool
	ExpectedFileSpec FileSpec
	ExpectedJSON     string
}

func TestFileSpecValidateBase(test *testing.T) {
	// Note that the expected file spec will not be validated (it should be constructed valid).
	testCases := []struct {
		spec           *FileSpec
		onlyLocalPaths bool
		expected       *FileSpec
		errorSubstring string
	}{
		// Base
		{
			&FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_NIL,
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_NIL,
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
				Dest: " ",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "/b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "/b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "a/b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "a/b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/a.zip",
				Dest: "b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/a.zip",
				Dest: "b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "b.zip",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "b.zip",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "/b.zip",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "/b.zip",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "a/b.zip",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "a/b.zip",
			},
			"",
		},
		// Casing
		{
			&FileSpec{
				Type: "EmPtY",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
			},
			"",
		},
		// Normalize Paths
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: " a  	",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/../b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a//b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "./a",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/a",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/..",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: ".",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: ".",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: ".",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "..",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "..",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "../a",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "../a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/../..",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "..",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/../../b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "../b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a/b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a/b",
				Dest: "b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "a/b/..",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "a.git",
				Dest: "a",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "a.zip",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a/b",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a/b",
				Dest: "b",
			},
			"",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "a/b/..",
			},
			false,
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "a.zip",
				Dest: "a",
			},
			"",
		},

		// Errors

		// Empty
		{
			nil,
			false,
			nil,
			"File spec is nil",
		},
		{
			&FileSpec{},
			false,
			nil,
			"Unknown FileSpec type",
		},
		// Invalid Contents
		{
			&FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
				Path: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
				Dest: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:      FILESPEC_TYPE_EMPTY,
				Reference: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:     FILESPEC_TYPE_EMPTY,
				Username: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:  FILESPEC_TYPE_EMPTY,
				Token: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_NIL,
				Path: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_NIL,
				Dest: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:      FILESPEC_TYPE_NIL,
				Reference: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:     FILESPEC_TYPE_NIL,
				Username: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type:  FILESPEC_TYPE_NIL,
				Token: "a",
			},
			false,
			nil,
			"An empty/nil FileSpec should have no other fields set",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
			},
			false,
			nil,
			"path cannot be empty",
		},
		{
			&FileSpec{
				Type:      FILESPEC_TYPE_PATH,
				Path:      "a",
				Reference: "b",
			},
			false,
			nil,
			"should not have reference, username, or token fields set",
		},
		{
			&FileSpec{
				Type:     FILESPEC_TYPE_PATH,
				Path:     "a",
				Username: "b",
			},
			false,
			nil,
			"should not have reference, username, or token fields set",
		},
		{
			&FileSpec{
				Type:  FILESPEC_TYPE_PATH,
				Path:  "a",
				Token: "b",
			},
			false,
			nil,
			"should not have reference, username, or token fields set",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_GIT,
			},
			false,
			nil,
			"cannot have an empty path",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_URL,
			},
			false,
			nil,
			"cannot have an empty path",
		},
		// Path Errors
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "",
			},
			false,
			nil,
			"cannot be empty",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/a",
			},
			true,
			nil,
			"not allowed to be absolute",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "/",
			},
			true,
			nil,
			"not allowed to be absolute",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "..",
			},
			true,
			nil,
			"points outside of the its base directory",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "../a",
			},
			true,
			nil,
			"points outside of the its base directory",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/../..",
			},
			true,
			nil,
			"points outside of the its base directory",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a/../../b",
			},
			true,
			nil,
			"points outside of the its base directory",
		},
		{
			&FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: "a",
				Dest: "/b",
			},
			true,
			nil,
			"not allowed to be absolute",
		},
		// Unknown Type
		{
			&FileSpec{
				Type: "ZZZ",
			},
			false,
			nil,
			"Unknown FileSpec type",
		},
	}

	for i, testCase := range testCases {
		err := testCase.spec.ValidateFull(testCase.onlyLocalPaths)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate spec '%+v': '%v'.", i, testCase.spec, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error '%s'.", i, MustToJSONIndent(testCase.spec))
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.spec) {
			test.Errorf("Case %d: Spec not as expected. Expected: '%s', Actual: '%s'.",
				i, MustToJSONIndent(testCase.expected), MustToJSONIndent(testCase.spec))
			continue
		}
	}
}

func TestFileSpecParseValidation(test *testing.T) {
	for i, testCase := range testCasesParseValidation {
		var spec FileSpec
		err := JSONFromString(testCase.Text, &spec)

		if testCase.Parses && (err != nil) {
			test.Errorf("Case %d: Could not unmarshal FileSpec '%s': '%v'.", i, testCase.Text, err)
			continue
		} else if !testCase.Parses && (err == nil) {
			test.Errorf("Case %d: FileSpec parsed, when it should not '%s'.", i, testCase.Text)
			continue
		}

		if !testCase.Parses {
			continue
		}

		err = spec.Validate()
		if testCase.Validates && (err != nil) {
			test.Errorf("Case %d: Failed to validate when it should '%+v': '%v'.", i, spec, err)
			continue
		} else if !testCase.Validates && (err == nil) {
			test.Errorf("Case %d: Validated when it should have failed '%+v'.", i, spec)
			continue
		}

		if !testCase.Validates {
			continue
		}

		if testCase.ExpectedFileSpec != spec {
			test.Errorf("Case %d: FileSpec is not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.ExpectedFileSpec, spec)
			continue
		}

		jsonString, err := ToJSON(spec)
		if err != nil {
			test.Errorf("Case %d: Could not marshal FileSpec '%+v': '%v'.", i, spec, err)
			continue
		}

		if testCase.ExpectedJSON != jsonString {
			test.Errorf("Case %d: FileSpec JSON is not as expected. Expected: '%s', Actual: '%s'.", i, testCase.ExpectedJSON, jsonString)
			continue
		}
	}
}

type testCaseCopy struct {
	Spec                  FileSpec
	ExpectedCopiedDirents []string
	ExpectErrorForSameDir bool
}

func TestFileSpecCopy(test *testing.T) {
	for i, testCase := range getCopyTestCases() {
		err := testCase.Spec.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate (%+v): '%v'.", i, testCase.Spec, err)
			continue
		}

		tempDir, err := MkDirTemp("autograder-test-filespec-copy-")
		if err != nil {
			test.Errorf("Case %d: Failed to create temp dir: '%v'.", i, err)
			continue
		}
		defer RemoveDirent(tempDir)

		destDir := filepath.Join(tempDir, "dest")
		if testCase.ExpectErrorForSameDir {
			destDir = filepath.Dir(testCase.Spec.Path)
		} else {
			MustMkDir(destDir)
		}

		err = testCase.Spec.CopyTarget(TestdataDirForTesting(), destDir)
		if (!testCase.ExpectErrorForSameDir) && (err != nil) {
			test.Errorf("Case %d: Failed to copy matching targets (%+v): '%v'.", i, testCase.Spec, err)
			continue
		} else if (testCase.ExpectErrorForSameDir) && (err == nil) {
			test.Errorf("Case %d: Unexpectedly copied matching targets (%+v).", i, testCase.Spec)
		}

		if testCase.ExpectErrorForSameDir {
			continue
		}

		dirents, err := GetAllDirents(destDir)
		if err != nil {
			test.Errorf("Case %d: Failed to get all dirents: '%v'.", i, err)
			continue
		}

		copiedDirents := []string{}
		for _, dirent := range dirents {
			relativeDirentPath, err := filepath.Rel(destDir, dirent)
			if err != nil {
				test.Errorf("Case %d: Failed to compute relative path for '%s': '%v'.", i, dirent, err)
				continue
			}
			copiedDirents = append(copiedDirents, relativeDirentPath)
		}

		if !reflect.DeepEqual(testCase.ExpectedCopiedDirents, copiedDirents) {
			test.Errorf("Case %d: Unexpected dirents copied. Expected: '%v', Actual: '%v'.", i, testCase.ExpectedCopiedDirents, copiedDirents)
			continue
		}
	}
}

func TestFileSpecDestPath(test *testing.T) {
	testCases := []struct {
		spec         FileSpec
		destDir      string
		expectedDest string
	}{
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "",
			},
			destDir:      "test",
			expectedDest: "test",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "/a.txt",
			},
			destDir:      "test",
			expectedDest: "/a.txt",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "a.txt",
			},
			destDir:      "test",
			expectedDest: "test/a.txt",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "",
			},
			destDir:      "",
			expectedDest: "",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "/a.txt",
			},
			destDir:      "",
			expectedDest: "/a.txt",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_PATH,
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
				Dest: "a.txt",
			},
			destDir:      "",
			expectedDest: "a.txt",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_EMPTY,
			},
			destDir:      "test",
			expectedDest: "test",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_NIL,
			},
			destDir:      "test",
			expectedDest: "test",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "",
			},
			destDir:      "",
			expectedDest: "foo",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "",
			},
			destDir:      "test",
			expectedDest: "test/foo",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "bar",
			},
			destDir:      "",
			expectedDest: "bar",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "bar",
			},
			destDir:      "test",
			expectedDest: "test/bar",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "/bar",
			},
			destDir:      "",
			expectedDest: "/bar",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_GIT,
				Path: "http://github.com/edulinq/foo.git",
				Dest: "/bar",
			},
			destDir:      "test",
			expectedDest: "/bar",
		},

		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "",
			},
			destDir:      "",
			expectedDest: "abc.zip",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "",
			},
			destDir:      "test",
			expectedDest: "test/abc.zip",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "bar.zip",
			},
			destDir:      "",
			expectedDest: "bar.zip",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "bar.zip",
			},
			destDir:      "test",
			expectedDest: "test/bar.zip",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "/bar.zip",
			},
			destDir:      "",
			expectedDest: "/bar.zip",
		},
		{
			spec: FileSpec{
				Type: FILESPEC_TYPE_URL,
				Path: "http://test.edulinq.org/abc.zip",
				Dest: "/bar.zip",
			},
			destDir:      "test",
			expectedDest: "/bar.zip",
		},
	}

	for i, testCase := range testCases {
		err := testCase.spec.Validate()
		if err != nil {
			test.Errorf("Case %d: Spec does not validate: '%v'.", i, err)
			continue
		}

		actualDest := testCase.spec.GetDest(testCase.destDir)
		if testCase.expectedDest != actualDest {
			test.Errorf("Case %d: Unexpected dest. Expected: '%s', Actual: '%s'.", i, testCase.expectedDest, actualDest)
			continue
		}
	}
}

// Note that this needs to be a function instead of a variable since the testing base dir does not get set until after static init.
func getCopyTestCases() []*testCaseCopy {
	return []*testCaseCopy{
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spec.txt"),
			},
			ExpectedCopiedDirents: []string{"spec.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test"),
			},
			ExpectedCopiedDirents: []string{"filespec_test", "filespec_test/spec.txt", "filespec_test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "*"),
			},
			ExpectedCopiedDirents: []string{"spec.txt", "spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "spe?.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "*_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/*globSpec.txt", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "*_test", "*"),
			},
			ExpectedCopiedDirents: []string{"*globSpec.txt", "spec.txt", "spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "f*_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		// Only one file is matched, so it will be renamed.
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", `\**_test`, "*"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "*"),
				Dest: "spec.txt",
			},
			ExpectErrorForSameDir: true,
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "*.txt"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "f*_test", "*.txt"),
				Dest: "test.test",
			},
			ExpectedCopiedDirents: []string{"test.test", "test.test/spec.txt", "test.test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", `\*globFileSpec_test`, `\*globSpec.txt`),
				Dest: `\*test.txt`,
			},
			ExpectedCopiedDirents: []string{`\*test.txt`},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "file????_test", "*"),
			},
			ExpectedCopiedDirents: []string{"spec.txt", "spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "????.txt"),
				Dest: "test.test",
			},
			ExpectedCopiedDirents: []string{"test.test"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "file????_test", "????.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespe[b-d]_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespe[^d-z]_test", "*"),
				Dest: "test",
			},
			ExpectedCopiedDirents: []string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "[r-t]pec.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespec_test", "[^a-r^t-v]pec.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespe[b-d]_test", "[r-t]pec.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
		&testCaseCopy{
			Spec: FileSpec{
				Type: "path",
				Path: filepath.Join(TestdataDirForTesting(), "files", "filespe[^d-z]_test", "[^a-r]pec.txt"),
				Dest: "test.txt",
			},
			ExpectedCopiedDirents: []string{"test.txt"},
		},
	}
}

var testCasesParseValidation []*testCaseParseValidation = []*testCaseParseValidation{
	// Empty.
	&testCaseParseValidation{
		`""`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_EMPTY},
		`{"type":"empty"}`,
	},
	&testCaseParseValidation{
		`null`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_EMPTY},
		`{"type":"empty"}`,
	},
	&testCaseParseValidation{
		`{"type": "empty"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_EMPTY},
		`{"type":"empty"}`,
	},

	// Nil.
	&testCaseParseValidation{
		`{"type": "nil"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_NIL},
		`{"type":"nil"}`,
	},

	// Path.
	&testCaseParseValidation{
		`"some/path"`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_PATH, Path: "some/path"},
		`{"type":"path","path":"some/path"}`,
	},
	&testCaseParseValidation{
		`{"type": "path", "path": "some/path"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_PATH, Path: "some/path"},
		`{"type":"path","path":"some/path"}`,
	},
	&testCaseParseValidation{
		`{"type": "path", "path": "some/path", "dest": "dirname"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_PATH, Path: "some/path", Dest: "dirname"},
		`{"type":"path","path":"some/path","dest":"dirname"}`,
	},
	&testCaseParseValidation{
		`{"type": "path", "path": "some/path/*.txt"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_PATH, Path: "some/path/*.txt"},
		`{"type":"path","path":"some/path/*.txt"}`,
	},

	// Git.
	&testCaseParseValidation{
		`{"type": "git", "path": "http://github.com/foo/bar.git"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_GIT, Path: "http://github.com/foo/bar.git", Dest: "bar"},
		`{"type":"git","path":"http://github.com/foo/bar.git","dest":"bar"}`,
	},
	&testCaseParseValidation{
		`{"type": "git", "path": "http://github.com/foo/bar.git", "dest": "baz"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_GIT, Path: "http://github.com/foo/bar.git", Dest: "baz"},
		`{"type":"git","path":"http://github.com/foo/bar.git","dest":"baz"}`,
	},
	&testCaseParseValidation{
		`{"type": "git", "path": "http://github.com/foo/bar.git", "dest": "baz", "reference": "main"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_GIT, Path: "http://github.com/foo/bar.git", Dest: "baz", Reference: "main"},
		`{"type":"git","path":"http://github.com/foo/bar.git","dest":"baz","reference":"main"}`,
	},
	&testCaseParseValidation{
		`{"type": "git", "path": "http://github.com/foo/bar.git", "dest": "baz", "reference": "main", "username": "user"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_GIT, Path: "http://github.com/foo/bar.git", Dest: "baz", Reference: "main", Username: "user"},
		`{"type":"git","path":"http://github.com/foo/bar.git","dest":"baz","reference":"main","username":"user"}`,
	},
	&testCaseParseValidation{
		`{"type": "git", "path": "http://github.com/foo/bar.git", "dest": "baz", "reference": "main", "username": "user", "token": "pass"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_GIT, Path: "http://github.com/foo/bar.git", Dest: "baz", Reference: "main", Username: "user", Token: "pass"},
		`{"type":"git","path":"http://github.com/foo/bar.git","dest":"baz","reference":"main","username":"user","token":"pass"}`,
	},

	// URL.
	&testCaseParseValidation{
		`"http://test.edulinq.org/abc.zip"`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://test.edulinq.org/abc.zip", Dest: "abc.zip"},
		`{"type":"url","path":"http://test.edulinq.org/abc.zip","dest":"abc.zip"}`,
	},
	&testCaseParseValidation{
		`{"type": "url", "path": "http://test.edulinq.org/abc.zip"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://test.edulinq.org/abc.zip", Dest: "abc.zip"},
		`{"type":"url","path":"http://test.edulinq.org/abc.zip","dest":"abc.zip"}`,
	},
	&testCaseParseValidation{
		`{"type": "url", "path": "http://test.edulinq.org/abc.zip", "dest": "xyz.txt"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://test.edulinq.org/abc.zip", Dest: "xyz.txt"},
		`{"type":"url","path":"http://test.edulinq.org/abc.zip","dest":"xyz.txt"}`,
	},

	// Validate failures.

	&testCaseParseValidation{
		`{"type": "empty", "path": "some/path"}`,
		true, false,
		FileSpec{},
		"",
	},

	&testCaseParseValidation{
		`{"type": "nil", "path": "some/path"}`,
		true, false,
		FileSpec{},
		"",
	},

	&testCaseParseValidation{
		`{"type": "path", "path": "some/[path"}`,
		true, false,
		FileSpec{},
		"",
	},

	&testCaseParseValidation{
		`{"type": "path"}`,
		true, false,
		FileSpec{},
		"",
	},

	&testCaseParseValidation{
		`{"type": "git"}`,
		true, false,
		FileSpec{},
		"",
	},
}

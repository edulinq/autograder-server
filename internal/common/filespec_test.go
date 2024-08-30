package common

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

type testCaseParseValidation struct {
	Text             string
	Parses           bool
	Validates        bool
	ExpectedFileSpec FileSpec
	ExpectedJSON     string
}

func TestFileSpecParseValidation(test *testing.T) {
	for i, testCase := range testCasesParseValidation {
		var spec FileSpec
		err := util.JSONFromString(testCase.Text, &spec)

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

		jsonString, err := util.ToJSON(spec)
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
	Spec         FileSpec
	OnlyContents bool
	ResultMD5    string
}

func TestFileSpecCopy(test *testing.T) {
	for i, testCase := range testCasesCopy {
		err := testCase.Spec.Validate()
		if err != nil {
			test.Errorf("Case %d: Failed to validate (%+v): '%v'.", i, testCase.Spec, err)
			continue
		}

		tempDir, err := util.MkDirTemp("autograder-test-filespec-copy-")
		if err != nil {
			test.Errorf("Case %d: Failed to create temp dir: '%v'.", i, err)
			continue
		}
		defer util.RemoveDirent(tempDir)

		destDir := filepath.Join(tempDir, "dest")

		err = testCase.Spec.CopyTarget(config.GetCourseImportDir(), destDir, testCase.OnlyContents)
		if err != nil {
			test.Errorf("Case %d: Failed to copy target (%+v): '%v'.", i, testCase.Spec, err)
			continue
		}

		zipPath := filepath.Join(tempDir, "contents.zip")
		err = util.Zip(destDir, zipPath, true)
		if err != nil {
			test.Errorf("Case %d: Failed to create zip file: '%v'.", i, err)
			continue
		}

		md5, err := util.MD5FileHex(zipPath)
		if err != nil {
			test.Errorf("Case %d: Failed to get zip's MD5 hash: '%v'.", i, err)
			continue
		}

		if testCase.ResultMD5 != md5 {
			test.Errorf("Case %d: MD5 mismatch. Expected: '%s', Actual: '%s'.", i, testCase.ResultMD5, md5)
			continue
		}
	}
}

var testCasesCopy []*testCaseCopy = []*testCaseCopy{
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test", "spec.txt")},
		false,
		"4b6070442f731cb7a83b18e0145c6be1",
	},
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test")},
		false,
		"7bf405075ec83eec2c44ac44ce3385c2",
	},
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test")},
		true,
		"4b6070442f731cb7a83b18e0145c6be1",
	},
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test", "spec.txt"), Dest: "test.txt"},
		false,
		"6911edb915da8cd1fdf1e4b1483d604e",
	},
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test"), Dest: "test"},
		false,
		"720b5768bf364f35c316976e549f1bcd",
	},
	&testCaseCopy{
		FileSpec{Type: "path", Path: filepath.Join(util.RootDirForTesting(), "testdata", "files", "filespec_test"), Dest: "test"},
		true,
		"720b5768bf364f35c316976e549f1bcd",
	},
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
		`{"type": "url", "path": "http://www.test.edulinq.org/abc.zip"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://www.test.edulinq.org/abc.zip", Dest: "abc.zip"},
		`{"type":"url","path":"http://www.test.edulinq.org/abc.zip","dest":"abc.zip"}`,
	},
	&testCaseParseValidation{
		`{"type": "url", "path": "http://www.test.edulinq.org/abc.zip", "dest": "xyz.txt"}`,
		true, true,
		FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://www.test.edulinq.org/abc.zip", Dest: "xyz.txt"},
		`{"type":"url","path":"http://www.test.edulinq.org/abc.zip","dest":"xyz.txt"}`,
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

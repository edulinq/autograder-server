package common

import (
	"path/filepath"
	"reflect"
	"strings"
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
	Spec                  FileSpec
	OnlyContents          bool
	ExpectedCopiedDirents []string
}

const PATH_SEPARATOR = "dest/"

func TestFileSpecCopy(test *testing.T) {
	for i, testCase := range getCopyTestCases() {
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

		err = testCase.Spec.CopyTarget(config.GetTestdataDir(), destDir, testCase.OnlyContents)
		if err != nil {
			test.Errorf("Case %d: Failed to copy target (%+v): '%v'.", i, testCase.Spec, err)
			continue
		}

		dirents, err := util.GetAllDirents(destDir)
		if err != nil {
			test.Errorf("Case %d: Failed to get all dirents: '%v'.", i, err)
			continue
		}

		copiedDirents := []string{}
		for _, dirent := range dirents {
			parts := strings.SplitN(dirent, PATH_SEPARATOR, 2)
			if len(parts) != 2 {
				test.Errorf("Case %d: Failed to split the dirent '%v'.", i, parts)
			}
			copiedDirents = append(copiedDirents, parts[1])
		}

		if !reflect.DeepEqual(testCase.ExpectedCopiedDirents, copiedDirents) {
			test.Errorf("Case %d: Unexpected contents copied. Expected: '%v', Actual: '%v'.", i, testCase.ExpectedCopiedDirents, copiedDirents)
			continue
		}
	}
}

// Note that this needs to be a function instead of a variable since the testing base dir does not get set until after static init.
func getCopyTestCases() []*testCaseCopy {
	return []*testCaseCopy{
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "spec.txt")},
			false,
			[]string{"spec.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test")},
			false,
			[]string{"filespec_test", "filespec_test/spec.txt", "filespec_test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test")},
			true,
			[]string{"spec.txt", "spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "spec.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test"), Dest: "test"},
			false,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test"), Dest: "test"},
			true,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "*_test"), Dest: "test"},
			true,
			[]string{"test", "test/*globSpec.txt", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "f*_test"), Dest: "test"},
			true,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", `\**_test`), Dest: "test"},
			true,
			[]string{"test", "test/*globSpec.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "*")},
			false,
			[]string{"spec.txt", "spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "f*_test", "*.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", `\*globFileSpec_test`, `\*globSpec.txt`), Dest: `\*test.txt`},
			false,
			[]string{`\*test.txt`},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "file????_test"), Dest: "test"},
			true,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "????.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "file????_test", "????.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespe[b-d]_test"), Dest: "test"},
			true,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespe[^d-z]_test"), Dest: "test"},
			true,
			[]string{"test", "test/spec.txt", "test/spec2.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "[r-t]pec.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespec_test", "[^a-r^t-v]pec.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespe[b-d]_test", "[r-t]pec.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
		},
		&testCaseCopy{
			FileSpec{Type: "path", Path: filepath.Join(config.GetTestdataDir(), "files", "filespe[^d-z]_test", "[^a-r]pec.txt"), Dest: "test.txt"},
			false,
			[]string{"test.txt"},
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

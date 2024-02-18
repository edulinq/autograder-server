package common

import (
    "path/filepath"
    "testing"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/util"
)

type testCaseParseValidation struct {
    Text string
    Parses bool
    Validates bool
    ExpectedFileSpec FileSpec
    ExpectedJSON string
}

func TestFileSpecParseValidation(test *testing.T) {
    for i, testCase := range testCasesParseValidation {
        var spec FileSpec;
        err := util.JSONFromString(testCase.Text, &spec);

        if (testCase.Parses && (err != nil)) {
            test.Errorf("Case %d: Could not unmarshal FileSpec '%s': '%v'.", i, testCase.Text, err);
            continue;
        } else if (!testCase.Parses && (err == nil)) {
            test.Errorf("Case %d: FileSpec parsed, when it should not '%s'.", i, testCase.Text);
            continue;
        }

        if (!testCase.Parses) {
            continue;
        }

        err = spec.Validate();
        if (testCase.Validates && (err != nil)) {
            test.Errorf("Case %d: Failed to validate when it should '%+v': '%v'.", i, spec, err);
            continue;
        } else if (!testCase.Validates && (err == nil)) {
            test.Errorf("Case %d: Validated when it should have failed '%+v'.", i, spec);
            continue;
        }

        if (!testCase.Validates) {
            continue;
        }

        if (testCase.ExpectedFileSpec != spec) {
            test.Errorf("Case %d: FileSpec is not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.ExpectedFileSpec, spec);
            continue;
        }

        jsonString, err := util.ToJSON(spec);
        if (err != nil) {
            test.Errorf("Case %d: Could not marshal FileSpec '%+v': '%v'.", i, spec, err);
            continue;
        }

        if (testCase.ExpectedJSON != jsonString) {
            test.Errorf("Case %d: FileSpec JSON is not as expected. Expected: '%s', Actual: '%s'.", i, testCase.ExpectedJSON, jsonString);
            continue;
        }
    }
}

type testCaseCopy struct {
    Spec FileSpec
    OnlyContents bool
    ResultMD5 string
}

func TestFileSpecCopy(test *testing.T) {
    for i, testCase := range testCasesCopy {
        err := testCase.Spec.Validate();
        if (err != nil) {
            test.Errorf("Case %d: Failed to validate (%+v): '%v'.", i, testCase.Spec, err);
            continue;
        }

        tempDir, err := util.MkDirTemp("autograder-test-filespec-copy-");
        if (err != nil) {
            test.Errorf("Case %d: Failed to create temp dir: '%v'.", i, err);
            continue;
        }
        defer util.RemoveDirent(tempDir);

        destDir := filepath.Join(tempDir, "dest");

        err = testCase.Spec.CopyTarget(config.GetCourseImportDir(), destDir, testCase.OnlyContents);
        if (err != nil) {
            test.Errorf("Case %d: Failed to copy target (%+v): '%v'.", i, testCase.Spec, err);
            continue;
        }

        zipPath := filepath.Join(tempDir, "contents.zip");
        err = util.Zip(destDir, zipPath, true);
        if (err != nil) {
            test.Errorf("Case %d: Failed to create zip file: '%v'.", i, err);
            continue;
        }

        md5, err := util.MD5FileHex(zipPath);
        if (err != nil) {
            test.Errorf("Case %d: Failed to get zip's MD5 hash: '%v'.", i, err);
            continue;
        }

        if (testCase.ResultMD5 != md5) {
            test.Errorf("Case %d: MD5 mismatch. Expected: '%s', Actual: '%s'.", i, testCase.ResultMD5, md5);
            continue;
        }
    }
}

var testCasesCopy []*testCaseCopy = []*testCaseCopy{
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files/a.txt"},
        false,
        "1d81dcf2c1a5b659901389d694c57248",
    },
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files"},
        false,
        "3ec09caa093ac5b9211001adf722a418",
    },
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files"},
        true,
        "e7571feb3aee557546b047c009017a3b",
    },
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files/a.txt", Dest: "test.txt"},
        false,
        "6911edb915da8cd1fdf1e4b1483d604e",
    },
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files", Dest: "test"},
        false,
        "1bb71f6af1335959240f755d148ce437",
    },
    &testCaseCopy{
        FileSpec{Type: "path", Path: "_tests/files", Dest: "test"},
        true,
        "1bb71f6af1335959240f755d148ce437",
    },
};

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
        `{"type": "url", "path": "http://www.test.com/abc.zip"}`,
        true, true,
        FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://www.test.com/abc.zip", Dest: "abc.zip"},
        `{"type":"url","path":"http://www.test.com/abc.zip","dest":"abc.zip"}`,
    },
    &testCaseParseValidation{
        `{"type": "url", "path": "http://www.test.com/abc.zip", "dest": "xyz.txt"}`,
        true, true,
        FileSpec{Type: FILESPEC_TYPE_URL, Path: "http://www.test.com/abc.zip", Dest: "xyz.txt"},
        `{"type":"url","path":"http://www.test.com/abc.zip","dest":"xyz.txt"}`,
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

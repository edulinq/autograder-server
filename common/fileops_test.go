package common

import (
    "path/filepath"
    "testing"

    "github.com/eriq-augustine/autograder/util"
)

func TestFileOpsToUnix(test *testing.T) {
    baseDir := "/tmp/test"

    testCases := []struct{Op FileOperation; Expected string}{
        {FileOperation([]string{"cp", "a", "b"}), "cp -r /tmp/test/a /tmp/test/b"},
        {FileOperation([]string{"mv", "a", "b"}), "mv /tmp/test/a /tmp/test/b"},

        {FileOperation([]string{"CP", "a", "b"}), "cp -r /tmp/test/a /tmp/test/b"},
        {FileOperation([]string{"MV", "a", "b"}), "mv /tmp/test/a /tmp/test/b"},

        {FileOperation([]string{"cp", "/a", "/b"}), "cp -r /a /b"},
        {FileOperation([]string{"mv", "/a", "/b"}), "mv /a /b"},

        {FileOperation([]string{"cp", "a A", "b B"}), "cp -r '/tmp/test/a A' '/tmp/test/b B'"},
        {FileOperation([]string{"mv", "a A", "b B"}), "mv '/tmp/test/a A' '/tmp/test/b B'"},

        {FileOperation([]string{"cp", "\"a\"", "'b'"}), "cp -r '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
        {FileOperation([]string{"mv", "\"a\"", "'b'"}), "mv '/tmp/test/\"a\"' '/tmp/test/'\"'\"'b'\"'\"''"},
    };

    for i, testCase := range testCases {
        err := testCase.Op.Validate();
        if (err != nil) {
            test.Errorf("Case %d: Failed to validate operation '%+v': '%v'.", i, testCase.Op, err);
            continue;
        }

        actual := testCase.Op.ToUnix(baseDir);
        if (testCase.Expected != actual) {
            test.Errorf("Case %d: Unexpected UNIX command. Expected `%s`, Actual: `%s`.", i, testCase.Expected, actual);
            continue;
        }
    }
}

func TestFileOpsCopyBase(test *testing.T) {
    tempDir := prepTempDir(test);
    defer util.RemoveDirent(tempDir);

    op := FileOperation([]string{"cp", "a.txt", "b.txt"});

    err := op.Validate();
    if (err != nil) {
        test.Fatalf("Failed to validate: '%v'.", err);
    }

    err = op.Exec(tempDir);
    if (err != nil) {
        test.Fatalf("Failed to exec: '%v'.", err);
    }

    // Getting the hash afer the operation will ensure the file still exists.
    expectedHash, err := util.MD5FileHex(filepath.Join(tempDir, "a.txt"));
    if (err != nil) {
        test.Fatalf("Failed to get expected hash: '%v'.", err);
    }

    actualHash, err := util.MD5FileHex(filepath.Join(tempDir, "b.txt"));
    if (err != nil) {
        test.Fatalf("Failed to get actual hash: '%v'.", err);
    }

    if (expectedHash != actualHash) {
        test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash);
    }
}

func TestFileOpsMoveBase(test *testing.T) {
    tempDir := prepTempDir(test);
    defer util.RemoveDirent(tempDir);

    op := FileOperation([]string{"mv", "a.txt", "b.txt"});

    err := op.Validate();
    if (err != nil) {
        test.Fatalf("Failed to validate: '%v'.", err);
    }

    expectedHash, err := util.MD5FileHex(filepath.Join(tempDir, "a.txt"));
    if (err != nil) {
        test.Fatalf("Failed to get expected hash: '%v'.", err);
    }

    err = op.Exec(tempDir);
    if (err != nil) {
        test.Fatalf("Failed to exec: '%v'.", err);
    }

    if (util.PathExists(filepath.Join(tempDir, "a.txt"))) {
        test.Fatalf("Source of move still exists.");
    }

    actualHash, err := util.MD5FileHex(filepath.Join(tempDir, "b.txt"));
    if (err != nil) {
        test.Fatalf("Failed to get actual hash: '%v'.", err);
    }

    if (expectedHash != actualHash) {
        test.Fatalf("Hashes to not match. Expected: '%s', Actual: '%s'.", expectedHash, actualHash);
    }
}

func prepTempDir(test *testing.T) string {
    tempDir, err := util.MkDirTemp("autograder-testing-fileop-");
    if (err != nil) {
        test.Fatalf("Failed to create temp dir: '%v'.", err);
    }

    err = util.WriteFile("AAA\n", filepath.Join(tempDir, "a.txt"));
    if (err != nil) {
        test.Fatalf("Failed to write test file: '%v'.", err);
    }

    return tempDir;
}

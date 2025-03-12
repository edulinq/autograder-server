package util

import (
	"path/filepath"
	"testing"
)

func TestPathHasParentOrSelfBase(test *testing.T) {
	testDir := MustMkDirTemp("test-realpathhasparentorself-")
	defer RemoveDirent(testDir)

	makeNestedTestDirs(testDir)

	testCases := []struct {
		relPath   string
		relParent string
		expected  bool
	}{
		// Self
		{
			"",
			"",
			true,
		},
		{
			".",
			".",
			true,
		},
		{
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "b", "c"),
			true,
		},
		{
			filepath.Join("a_symlink"),
			filepath.Join("a_symlink"),
			true,
		},

		// Simple Parents
		{
			"a",
			".",
			true,
		},
		{
			filepath.Join("a", "b"),
			"a",
			true,
		},
		{
			filepath.Join("a", "b", "c"),
			"a",
			true,
		},
		{
			filepath.Join("a", "b", "c"),
			filepath.Join("a", "b"),
			true,
		},

		// Simple Not Parents
		{
			".",
			filepath.Join("a"),
			false,
		},
		{
			filepath.Join("a", "b"),
			filepath.Join("a", "b", "c"),
			false,
		},

		// Successful Symbolic Links
		{
			filepath.Join("a_symlink"),
			".",
			true,
		},
		{
			filepath.Join("a", "b"),
			filepath.Join("a_symlink"),
			true,
		},
		{
			filepath.Join("b_symlink"),
			filepath.Join("a_symlink"),
			true,
		},
		{
			filepath.Join("c_symlink"),
			filepath.Join("a", "b"),
			true,
		},

		// Not Parent Symbolic Links
		{
			filepath.Join("a_symlink"),
			filepath.Join("b_symlink"),
			false,
		},
		{
			filepath.Join("a"),
			filepath.Join("b_symlink"),
			false,
		},
		{
			filepath.Join("a_symlink"),
			filepath.Join("b"),
			false,
		},

		// Not Exists
		{
			filepath.Join("a", "NOT_EXISTS"),
			filepath.Join("a"),
			true,
		},
		{
			filepath.Join("a", "b", "NOT_EXISTS"),
			filepath.Join("a"),
			true,
		},
		{
			filepath.Join("a"),
			filepath.Join("a", "NOT_EXISTS"),
			false,
		},
	}

	for i, testCase := range testCases {
		path := filepath.Join(testDir, testCase.relPath)
		parent := filepath.Join(testDir, testCase.relParent)

		actual := PathHasParentOrSelf(path, parent)
		if testCase.expected != actual {
			test.Errorf("Case %d: Incorrect result for ('%s', '%s'). Expected: '%v', Actual: '%v'.", i, testCase.relPath, testCase.relParent, testCase.expected, actual)
			continue
		}
	}
}

// mkdir -p a/b/c
//
// ln -s a     a_symlink
// ln -s a/b   b_symlink
// ln -s a/b/c c_symlink
func makeNestedTestDirs(baseDir string) {
	path := filepath.Join(baseDir, "a", "b", "c")
	MustMkDir(path)

	MustSymbolicLink(filepath.Join(baseDir, "a"), filepath.Join(baseDir, "a_symlink"))
	MustSymbolicLink(filepath.Join(baseDir, "a", "b"), filepath.Join(baseDir, "b_symlink"))
	MustSymbolicLink(filepath.Join(baseDir, "a", "b", "c"), filepath.Join(baseDir, "c_symlink"))
}

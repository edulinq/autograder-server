package util

import (
	"path/filepath"
	"testing"
)

func TestExtractPythonCodeFromNotebookFileBase(test *testing.T) {
	inputPath := filepath.Join(RootDirForTesting(), "testdata", "files", "python_notebook", "ipynb", "submission.ipynb")
	expectedPath := filepath.Join(RootDirForTesting(), "testdata", "files", "python_notebook", "py", "submission.py")

	expected := MustReadFile(expectedPath)

	actual, err := ExtractPythonCodeFromNotebookFile(inputPath)
	if err != nil {
		test.Fatalf("Failed to extract code: '%v'.", err)
	}

	if expected != actual {
		test.Fatalf("Result not as expected.\n--- expected ---\n%s\n---\n--- actual ---\n%s\n---", expected, actual)
	}
}

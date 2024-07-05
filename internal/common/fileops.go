package common

import (
	"errors"
	"fmt"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/edulinq/autograder/internal/util"
)

type FileOperation []string

// Copy over assignment filespecs.
// 1) Do pre-copy operations.
// 2) Copy.
// 3) Do post-copy operations.
func CopyFileSpecs(
	sourceDir string, destDir string, baseDir string,
	filespecs []*FileSpec, onlyContents bool,
	preOperations []FileOperation, postOperations []FileOperation) error {
	// Do pre ops.
	err := ExecFileOperations(preOperations, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to do pre file operation: '%w'.", err)
	}

	// Copy files.
	for _, filespec := range filespecs {
		err = filespec.CopyTarget(sourceDir, destDir, onlyContents)
		if err != nil {
			return fmt.Errorf("Failed to handle filespec '%s': '%w'", filespec, err)
		}
	}

	// Do post ops.
	err = ExecFileOperations(postOperations, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to do post file operation: '%w'.", err)
	}

	return nil
}

func (this FileOperation) String() string {
	return strings.Join([]string(this), " ")
}

func (this FileOperation) ToUnix(baseDir string) string {
	parts := []string(this)
	command := parts[0]

	var result []string

	if command == "cp" {
		sourcePath := resolvePath(parts[1], baseDir, true)
		destPath := resolvePath(parts[2], baseDir, true)

		result = []string{
			"cp",
			"-r",
			sourcePath,
			destPath,
		}
	} else if command == "mv" {
		sourcePath := resolvePath(parts[1], baseDir, true)
		destPath := resolvePath(parts[2], baseDir, true)

		result = []string{
			"mv",
			sourcePath,
			destPath,
		}
	} else {
		return fmt.Sprintf("echo 'Invalid FileOperation: \"%s\"'.", this.String())
	}

	return shellescape.QuoteCommand(result)
}

func (this FileOperation) Exec(baseDir string) error {
	parts := []string(this)
	command := parts[0]

	if command == "cp" {
		sourcePath := resolvePath(parts[1], baseDir, false)
		destPath := resolvePath(parts[2], baseDir, false)

		return util.CopyDirent(sourcePath, destPath, false)
	} else if command == "mv" {
		sourcePath := resolvePath(parts[1], baseDir, false)
		destPath := resolvePath(parts[2], baseDir, false)

		return os.Rename(sourcePath, destPath)
	} else {
		return fmt.Errorf("Unknown file operation: '%s'.", command)
	}
}

func resolvePath(path string, baseDir string, forceUnix bool) string {
	if !filepath.IsAbs(path) {
		if forceUnix {
			path = gopath.Join(baseDir, path)
		} else {
			path = filepath.Join(baseDir, path)
		}
	}

	return path
}

func (this FileOperation) Validate() error {
	parts := []string(this)

	if (this == nil) || (len(parts) == 0) {
		return fmt.Errorf("File operation is empty.")
	}

	parts[0] = strings.ToLower(parts[0])

	command := parts[0]
	length := len(parts)

	if command == "cp" {
		if length != 3 {
			return fmt.Errorf("Incorrect number of argument for 'cp' file operation. Expected 2, found %d.", length-1)
		}
	} else if command == "mv" {
		if length != 3 {
			return fmt.Errorf("Incorrect number of argument for 'mv' file operation. Expected 2, found %d.", length-1)
		}
	} else {
		return fmt.Errorf("Unknown file operation: '%s'.", command)
	}

	return nil
}

func ValidateFileOperations(operations []FileOperation) error {
	var errs error

	for _, operation := range operations {
		errors.Join(errs, operation.Validate())
	}

	return errs
}

func ExecFileOperations(operations []FileOperation, baseDir string) error {
	for _, operation := range operations {
		err := operation.Exec(baseDir)
		if err != nil {
			return fmt.Errorf("Failed to exec file operation '%s': '%w'.", operation.String(), err)
		}
	}

	return nil
}

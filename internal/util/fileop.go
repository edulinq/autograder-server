package util

import (
	"errors"
	"fmt"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"
)

// File operations represent simple file operations.
// Any represented file paths must be POSIX, relative, and not point to any parent directories.
// Note that this code will only work properly on POSIX systems because of the lexical analysis on paths.
type FileOperation []string

const (
	FILE_OP_LONG_COPY  = "copy"
	FILE_OP_SHORT_COPY = "cp"

	FILE_OP_LONG_MOVE  = "move"
	FILE_OP_SHORT_MOVE = "mv"
)

// The long name is the canonical name.
var fileOpNormalName map[string]string = map[string]string{
	FILE_OP_SHORT_COPY: FILE_OP_LONG_COPY,
	FILE_OP_LONG_COPY:  FILE_OP_LONG_COPY,

	FILE_OP_SHORT_MOVE: FILE_OP_LONG_MOVE,
	FILE_OP_LONG_MOVE:  FILE_OP_LONG_MOVE,
}

// The number of operations for each file operation.
var fileOpNumArgs map[string]int = map[string]int{
	FILE_OP_LONG_COPY: 2,
	FILE_OP_LONG_MOVE: 2,
}

func NewFileOperation(parts []string) *FileOperation {
	operation := FileOperation(parts)
	return &operation
}

func (this *FileOperation) Validate() error {
	if this == nil {
		return fmt.Errorf("File operation is nil.")
	}

	parts := []string(*this)
	if len(parts) == 0 {
		return fmt.Errorf("File operation is empty.")
	}

	command, ok := fileOpNormalName[strings.ToLower(parts[0])]
	if !ok {
		return fmt.Errorf("Unknown file operation: '%s'.", parts[0])
	}

	parts[0] = command

	numArgs := (len(parts) - 1)
	expectedNumArgs := fileOpNumArgs[command]
	if expectedNumArgs != numArgs {
		return fmt.Errorf("Incorrect number of arguments for '%s' file operation. Expected %d, found %d.", command, expectedNumArgs, numArgs)
	}

	// Check all path arguments.
	for i := 1; i < len(parts); i++ {
		path := parts[i]

		if strings.Contains(path, "\\") {
			return fmt.Errorf("Argument at index %d ('%s') contains a backslash ('\\') or is not a POSIX path.", i, parts[i])
		}

		path = filepath.Clean(path)

		if filepath.IsAbs(path) {
			return fmt.Errorf("Argument at index %d ('%s') is an absolute path. Only relative paths are allowed.", i, parts[i])
		}

		if !filepath.IsLocal(path) {
			return fmt.Errorf("Argument at index %d ('%s') points outside of the its base directory. File operation paths can not reference parent directories.", i, parts[i])
		}

		if path == "." {
			return fmt.Errorf("Argument at index %d ('%s') cannot point just to the current directory. File operation paths must point to a dirent inside the current directory tree.", i, parts[i])
		}

		parts[i] = path
	}

	return nil
}

func (this *FileOperation) String() string {
	return strings.Join([]string(*this), " ")
}

// Create a string that represents invoking this operation on a UNIX command-line.
func (this *FileOperation) ToUnix(baseDir string) string {
	parts := []string(*this)
	command := parts[0]

	var result []string

	if command == FILE_OP_LONG_COPY {
		sourcePath := resolvePath(parts[1], baseDir, true)
		destPath := resolvePath(parts[2], baseDir, true)

		result = []string{
			"cp",
			"-r",
			sourcePath,
			destPath,
		}
	} else if command == FILE_OP_LONG_MOVE {
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

// Execute this operation in the given directory.
func (this *FileOperation) Exec(baseDir string) error {
	parts := []string(*this)
	command := parts[0]

	if command == FILE_OP_LONG_COPY {
		sourcePath := resolvePath(parts[1], baseDir, false)
		destPath := resolvePath(parts[2], baseDir, false)

		return CopyDirent(sourcePath, destPath, false)
	} else if command == FILE_OP_LONG_MOVE {
		sourcePath := resolvePath(parts[1], baseDir, false)
		destPath := resolvePath(parts[2], baseDir, false)

		return os.Rename(sourcePath, destPath)
	} else {
		return fmt.Errorf("Unknown file operation: '%s'.", command)
	}
}

func ValidateFileOperations(operations []*FileOperation) error {
	var errs error

	for i, operation := range operations {
		if operation == nil {
			errors.Join(errs, fmt.Errorf("File operation at index %d is nil.", i))
			continue
		}

		err := operation.Validate()
		if err != nil {
			errors.Join(errs, fmt.Errorf("Failed to validate file operation at index %d: '%w'.", i, err))
			continue
		}
	}

	return errs
}

func ExecFileOperations(operations []*FileOperation, baseDir string) error {
	for _, operation := range operations {
		err := operation.Exec(baseDir)
		if err != nil {
			return fmt.Errorf("Failed to exec file operation '%s': '%w'.", operation.String(), err)
		}
	}

	return nil
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

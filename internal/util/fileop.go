package util

import (
	"errors"
	"fmt"
	gopath "path"
	"path/filepath"
	"strings"

	"al.essio.dev/pkg/shellescape"
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

	FILE_OP_LONG_MKDIR  = "make-dir"
	FILE_OP_SHORT_MKDIR = "mkdir"

	FILE_OP_LONG_REMOVE  = "remove"
	FILE_OP_SHORT_REMOVE = "rm"
)

// The long name is the canonical name.
var fileOpNormalName map[string]string = map[string]string{
	FILE_OP_SHORT_COPY: FILE_OP_LONG_COPY,
	FILE_OP_LONG_COPY:  FILE_OP_LONG_COPY,

	FILE_OP_SHORT_MOVE: FILE_OP_LONG_MOVE,
	FILE_OP_LONG_MOVE:  FILE_OP_LONG_MOVE,

	FILE_OP_SHORT_MKDIR: FILE_OP_LONG_MKDIR,
	FILE_OP_LONG_MKDIR:  FILE_OP_LONG_MKDIR,

	FILE_OP_SHORT_REMOVE: FILE_OP_LONG_REMOVE,
	FILE_OP_LONG_REMOVE:  FILE_OP_LONG_REMOVE,
}

// The number of operations for each file operation.
var fileOpNumArgs map[string]int = map[string]int{
	FILE_OP_LONG_COPY:   2,
	FILE_OP_LONG_MOVE:   2,
	FILE_OP_LONG_MKDIR:  1,
	FILE_OP_LONG_REMOVE: 1,
}

var supportsSourceGlobs map[string]bool = map[string]bool{
	FILE_OP_LONG_COPY:   true,
	FILE_OP_LONG_MOVE:   true,
	FILE_OP_LONG_MKDIR:  false,
	FILE_OP_LONG_REMOVE: true,
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
		path := strings.TrimSpace(parts[i])

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

	if supportsSourceGlobs[command] {
		_, err := filepath.Match(parts[1], "")
		if err != nil {
			return fmt.Errorf("Argument at index 1 ('%s') contains an invalid glob path: '%w'.", parts[1], err)
		}
	}

	return nil
}

func (this *FileOperation) String() string {
	return strings.Join([]string(*this), " ")
}

// Create a string that represents invoking this operation on a UNIX command-line.
// It is assumed that the commands output from here will only be run in a Docker container.
// Therefore, containment will not be checked (since it is already sandboxed).
func (this *FileOperation) ToUnixForDocker(baseDir string) string {
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
	} else if command == FILE_OP_LONG_MKDIR {
		path := resolvePath(parts[1], baseDir, true)

		result = []string{
			"mkdir",
			"-p",
			path,
		}
	} else if command == FILE_OP_LONG_REMOVE {
		path := resolvePath(parts[1], baseDir, true)

		result = []string{
			"rm",
			"-rf",
			path,
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

		return handleGlobFileOperation(sourcePath, destPath, baseDir, CopyDirent)
	} else if command == FILE_OP_LONG_MOVE {
		sourcePath := resolvePath(parts[1], baseDir, false)
		destPath := resolvePath(parts[2], baseDir, false)

		return handleGlobFileOperation(sourcePath, destPath, baseDir, MoveDirent)
	} else if command == FILE_OP_LONG_MKDIR {
		path := resolvePath(parts[1], baseDir, false)
		if !PathHasParentOrSelf(path, baseDir) {
			return fmt.Errorf("Target path breaks containment: '%s'.", path)
		}

		return MkDir(path)
	} else if command == FILE_OP_LONG_REMOVE {
		path := resolvePath(parts[1], baseDir, false)
		return handleGlobRemove(path, baseDir)
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
			path = ShouldNormalizePath(path)
		}
	}

	return path
}

func handleGlobFileOperation(sourceGlob string, dest string, baseDir string, operation func(string, string) error) error {
	if !PathHasParentOrSelf(dest, baseDir) {
		return fmt.Errorf("Dest path breaks containment: '%s'.", dest)
	}

	sourcePaths, err := prepForGlobs(sourceGlob, dest, baseDir)
	if err != nil {
		return fmt.Errorf("Failed to prep glob path '%s' for an operation: '%w'.", sourceGlob, err)
	}

	var errs error

	for _, sourcePath := range sourcePaths {
		realSourcePath, err := filepath.EvalSymlinks(sourcePath)
		if err != nil {
			return fmt.Errorf("Failed to eval symlinks for sourcePath '%s': '%w'.", sourcePath, err)
		}

		if realSourcePath == dest {
			continue
		}

		err = operation(sourcePath, dest)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func handleGlobRemove(path string, baseDir string) error {
	paths, err := filepath.Glob(path)
	if err != nil {
		return fmt.Errorf("Failed to resolve glob path '%s': '%w'.", path, err)
	}

	var errs error

	// Ensure that all the paths are contained.
	for _, path := range paths {
		if !PathHasParentOrSelf(path, baseDir) {
			errs = errors.Join(errs, fmt.Errorf("Target path breaks containment: '%s'.", path))
		}
	}

	if errs != nil {
		return errs
	}

	for _, path := range paths {
		err = RemoveDirent(path)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func prepForGlobs(globPath string, destPath string, baseDir string) ([]string, error) {
	sourcePaths, err := filepath.Glob(globPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve glob path '%s': '%w'.", globPath, err)
	}

	if len(sourcePaths) == 0 {
		return nil, fmt.Errorf("Unable to find source path: '%s'.", globPath)
	}

	// Ensure that all the paths are contained.
	var errs error
	for _, path := range sourcePaths {
		if !PathHasParentOrSelf(path, baseDir) {
			errs = errors.Join(errs, fmt.Errorf("Source path breaks containment: '%s'.", path))
		}
	}

	if errs != nil {
		return nil, errs
	}

	// If there are multiple source paths, dest path must be a directory.
	// This is to avoid multiple sources competing over the same destination file.
	if len(sourcePaths) > 1 {
		err = MkDir(destPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to create dest dir '%s': '%w'.", destPath, err)
		}
	}

	return sourcePaths, nil
}

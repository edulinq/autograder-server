package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func RemoveDirent(path string) error {
	return os.RemoveAll(path)
}

// Copy a file or directory into dest.
// If source is a file, then dest can be a file or dir.
// If source is a dir, then see CopyDir() for semantics.
func CopyDirent(source string, dest string, onlyContents bool) error {
	if !PathExists(source) {
		return fmt.Errorf("Source dirent for copy does not exist: '%s'", source)
	}

	if IsSymLink(source) {
		return CopyLink(source, dest)
	}

	if IsDir(source) {
		return CopyDir(source, dest, onlyContents)
	}

	return CopyFile(source, dest)
}

func CopyLink(source string, dest string) error {
	if !PathExists(source) {
		return fmt.Errorf("Source link for copy does not exist: '%s'", source)
	}

	if IsDir(dest) {
		dest = filepath.Join(dest, filepath.Base(source))
	}

	MkDir(filepath.Dir(dest))

	target, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("Failed to read link being copied (%s): %w.", source, err)
	}

	err = os.Symlink(target, dest)
	if err != nil {
		return fmt.Errorf("Failed to write link (target: '%s', source: '%s', dest: '%s'): %w.", target, source, dest, err)
	}

	return nil
}

// Copy a file from source to dest creating any necessary parents along the way.
// If dest is a file, it will be truncated.
// If dest is a dir, then the file will be created inside dest.
func CopyFile(source string, dest string) error {
	if !PathExists(source) {
		return fmt.Errorf("Source file for copy does not exist: '%s'", source)
	}

	if IsDir(dest) {
		dest = filepath.Join(dest, filepath.Base(source))
	}

	os.MkdirAll(filepath.Dir(dest), 0755)

	sourceIO, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Could not open source file for copy (%s): %w.", source, err)
	}
	defer sourceIO.Close()

	destIO, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Could not open dest file for copy (%s): %w.", dest, err)
	}
	defer destIO.Close()

	_, err = io.Copy(destIO, sourceIO)
	if err != nil {
		return fmt.Errorf("Failed to copy file contents from '%s' to '%s': %w.", source, dest, err)
	}

	return nil
}

// Copy a directory (or just it's contents) into dest.
// When onlyContents = False:
//   - dest must not exist.
//   - `cp -r source dest`
//
// When onlyContents = True:
//   - dest may exist (and must be a dir).
//   - `cp source/* dest/`
func CopyDir(source string, dest string, onlyContents bool) error {
	if onlyContents {
		return CopyDirContents(source, dest)
	}

	return CopyDirWhole(source, dest)
}

// Copy a directory (including it's contents) to a new path.
// Any non-existent parents will be created.
// dest must not exist.
func CopyDirWhole(source string, dest string) error {
	if !IsDir(source) {
		return fmt.Errorf("Source of whole directory copy ('%s') does not exist or is not a dir.", source)
	}

	if PathExists(dest) {
		return fmt.Errorf("Destination of whole directory copy ('%s') already exists.", dest)
	}

	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create destination of whole directory copy ('%s'): '%w'.", dest, err)
	}

	return CopyDirContents(source, dest)
}

// Copy the contents of source into dest.
// dest does not have to exist, but any conflicting files will be clobbered.
func CopyDirContents(source string, dest string) error {
	if !IsDir(source) {
		return fmt.Errorf("Source of directory copy ('%s') does not exist or is not a dir.", source)
	}

	if !PathExists(dest) {
		err := MkDir(dest)
		if err != nil {
			return fmt.Errorf("Failed to create dest dir '%s': '%w'.", dest, err)
		}
	}

	if !IsDir(dest) {
		return fmt.Errorf("Dest of directory copy ('%s') does not exist or is not a dir.", dest)
	}

	dirents, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("Could not list dir for copy '%s': '%w'.", source, err)
	}

	for _, dirent := range dirents {
		sourcePath := filepath.Join(source, dirent.Name())
		destPath := filepath.Join(dest, dirent.Name())

		err = CopyDirent(sourcePath, destPath, false)
		if err != nil {
			return err
		}
	}

	return nil
}

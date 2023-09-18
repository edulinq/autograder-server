package util

import (
    "fmt"
    "io"
    "path/filepath"
    "os"
)

// Copy a file or directory into dest.
// If source is a file, then dest can be a file or dir.
// If source is a dir, then see CopyDir() for semantics.
func CopyDirent(source string, dest string, onlyContents bool) error {
    if (!PathExists(source)) {
        return fmt.Errorf("Source dirent for copy does not exist: '%s'", source);
    }

    if (IsDir(source)) {
        return CopyDir(source, dest, onlyContents);
    }

    return CopyFile(source, dest);
}

// Copy a file from source to dest creating any necessary parents along the way..
// If dest is a file, it will be truncated.
// If fest is a dir, then the file will be created inside dest.
func CopyFile(source string, dest string) error {
    if (!PathExists(source)) {
        return fmt.Errorf("Source file for copy does not exist: '%s'", source);
    }

    if (IsDir(dest)) {
        dest = filepath.Join(dest, filepath.Base(source));
    }

    os.MkdirAll(filepath.Dir(dest), 0755);

    sourceIO, err := os.Open(source);
    if (err != nil) {
        return fmt.Errorf("Could not open source file for copy (%s): %w.", source, err);
    }
    defer sourceIO.Close()

    destIO, err := os.Create(dest);
    if (err != nil) {
        return fmt.Errorf("Could not open dest file for copy (%s): %w.", dest, err);
    }
    defer destIO.Close()

    _, err = io.Copy(destIO, sourceIO);
    return err;
}

// Copy a directory (or just it's contents) into dest.
// dest must not exist.
// When onlyContents = False:
//    - dest must not exist.
//    - `cp -r source dest`
// When onlyContents = True:
//    - dest must exist and be a dir.
//    - `cp source/* dest/`
func CopyDir(source string, dest string, onlyContents bool) error {
    if (onlyContents) {
        return CopyDirContents(source, dest);
    }

    return CopyDirWhole(source, dest);
}

// Copy a directory (including it's contents) to a new path.
// Any non-existent parents will be created.
// dest must not exist.
func CopyDirWhole(source string, dest string) error {
    if (!IsDir(source)) {
        return fmt.Errorf("Source of whole directory copy ('%s') does not exist or is not a dir.", source)
    }

    if (PathExists(dest)) {
        return fmt.Errorf("Destination of whole directory copy ('%s') already exists.", dest);
    }

    err := os.MkdirAll(dest, 0755);
    if (err != nil) {
        return fmt.Errorf("Failed to create destination of whole directory copy ('%s'): '%w'.", dest, err);
    }

    return CopyDirContents(source, dest);
}

// Copy the contents of source into dest.
// dest does not have to exist, but any conflicting files will be clobbered.
func CopyDirContents(source string, dest string) error {
    if (!IsDir(source)) {
        return fmt.Errorf("Source of directory copy ('%s') does not exist or is not a dir.", source)
    }

    dirents, err := os.ReadDir(source);
    if (err != nil) {
        return fmt.Errorf("Could not list dir for copy '%s': '%w'.", source, err);
    }

    for _, dirent := range dirents {
        sourcePath := filepath.Join(source, dirent.Name());
        destPath := filepath.Join(dest, dirent.Name());

        if (dirent.IsDir()) {
            err = CopyDirWhole(sourcePath, destPath);
            if (err != nil) {
                return err;
            }
        } else {
            err = CopyFile(sourcePath, destPath);
            if (err != nil) {
                return err;
            }
        }
    }

    return nil;
}

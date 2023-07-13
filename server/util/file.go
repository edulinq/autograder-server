package util

import (
    "fmt"
    "io"
    "path/filepath"
    "os"
)

func CopyFile(source string, dest string) error {
    if (!PathExists(source)) {
        return fmt.Errorf("Source file for copy does not exist: '%s'", source);
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

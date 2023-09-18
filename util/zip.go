package util

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

func Unzip(zipPath string, outDir string) error {
    reader, err := zip.OpenReader(zipPath);
    if (err != nil) {
        return fmt.Errorf("Could not open zip archive ('%s') for reading: '%w'.", zipPath, err);
    }
    defer reader.Close();

    for _, zipfile := range reader.File {
        path := filepath.Join(outDir, zipfile.Name);

        if (strings.HasSuffix(zipfile.Name, "/")) {
            // Dir
            os.MkdirAll(path, 0755);
            continue;
        }

        // File

        inFile, err := zipfile.Open();
        if (err != nil) {
            return fmt.Errorf("Could not open zip archive ('%s') file ('%s') for reading: '%w'.", zipPath, zipfile.Name, err);
        }
        defer inFile.Close();

        outFile, err := os.Create(path);
        if (err != nil) {
            return fmt.Errorf("Failed to create output file ('%s'): '%w'.", path, err);
        }
        defer outFile.Close();

        _, err = io.Copy(outFile, inFile);
        if (err != nil) {
            return fmt.Errorf("Could not write zip contents into file ('%s'): '%w'.", path, err);
        }

        inFile.Close();
        outFile.Close();
    }

    return nil;
}

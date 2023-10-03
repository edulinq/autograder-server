package util

import (
    "archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
)

func Zip(source string, dest string) error {
    if (!strings.HasSuffix(dest, ".zip")) {
        dest = dest + ".zip";
    }

    if (PathExists(dest)) {
        return fmt.Errorf("Path to zip target ('%s') already exists.", dest);
    }

    zipfile, err := os.Create(dest);
    if (err != nil) {
        return fmt.Errorf("Could not create file for zip target '%s': '%w'.", dest, err);
    }
    defer zipfile.Close();


    writer := zip.NewWriter(zipfile);
    defer writer.Close();

    err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
        if (err != nil) {
            return err;
        }

        header, err := zip.FileInfoHeader(info);
        if (err != nil) {
            return fmt.Errorf("Could not create file header for '%s': '%w'.", path, err);
        }

        header.Method = zip.Deflate;

        header.Name, err = filepath.Rel(filepath.Dir(source), path);
        if (err != nil) {
            return fmt.Errorf("Could not compute relative path for '%s' (wrt '%s'): '%w'.", path, source, err);
        }

        if (info.IsDir()) {
            header.Name += "/";
        }

        headerWriter, err := writer.CreateHeader(header);
        if (err != nil) {
            return fmt.Errorf("Could not create file header writer for '%s': '%w'.", path, err);
        }

        if (info.IsDir()) {
            return nil
        }

        file, err := os.Open(path);
        if (err != nil) {
            return fmt.Errorf("Could not create file handle for '%s': '%w'.", path, err);
        }
        defer file.Close();

        _, err = io.Copy(headerWriter, file);
        if (err != nil) {
            return fmt.Errorf("Could not copy file into zipfile '%s': '%w'.", path, err);
        }

        return nil;
    });

    if (err != nil) {
        return fmt.Errorf("Could not create zip file '%s' for source '%s': '%w'.", dest, source, err);
    }

    return nil;
}

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

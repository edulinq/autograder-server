package util

import (
    "archive/zip"
    "bytes"
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

    return AddDirToZipWriter(source, "", writer);
}

// Zip to a slice of bytes.
func ZipToBytes(source string, prefix string) ([]byte, error) {
    buffer := new(bytes.Buffer);

    // Create a new zip archive.
    writer := zip.NewWriter(buffer);
    defer writer.Close();

    err := AddDirToZipWriter(source, prefix, writer);
    if (err != nil) {
        return nil, err;
    }

    writer.Close();
    return buffer.Bytes(), nil;
}

// |prefix| can be used to set a dir that the zip contents will be located in.
func AddDirToZipWriter(source string, prefix string, writer *zip.Writer) error {
    err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
        if (err != nil) {
            return err;
        }

        archivePath, err := filepath.Rel(filepath.Dir(source), path);
        if (err != nil) {
            return fmt.Errorf("Could not compute relative path for '%s' (wrt '%s'): '%w'.", path, source, err);
        }

        if (prefix != "") {
            archivePath = filepath.Join(prefix, archivePath);
        }

        err = AddFileToZipWriter(path, archivePath, writer);
        if (err != nil) {
            return fmt.Errorf("Could not add file to zip archive '%s': '%w'.", path, err);
        }

        return nil;
    });

    if (err != nil) {
        return fmt.Errorf("Could not create zip file for source '%s': '%w'.", source, err);
    }

    return nil;
}

// Add an existing file to an ongoing zip writer.
// This can be used to add several files to an archive that may not start in the same source directory.
// |archivePath| is used to set the file's path/name within the archive.
func AddFileToZipWriter(path string, archivePath string, writer *zip.Writer) error {
    info, err := os.Stat(path);
    if (err != nil) {
        return fmt.Errorf("Could not stat source path '%s': '%w'.", path, err);
    }

    header, err := zip.FileInfoHeader(info);
    if (err != nil) {
        return fmt.Errorf("Could not create file header for '%s': '%w'.", path, err);
    }

    header.Method = zip.Deflate;
    header.Name = archivePath;

    if (info.IsDir() && !strings.HasSuffix(header.Name, "/")) {
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

// A type to help handling ongoing zip operations where files are added individually.
// Once closed or GetBytes() as been called, no new files can be added.
type OngoingZipOperation struct {
    buffer *bytes.Buffer
    writer *zip.Writer
}

func NewOngoingZipOperation() *OngoingZipOperation {
    buffer := new(bytes.Buffer);

    return &OngoingZipOperation{
        buffer: buffer,
        writer: zip.NewWriter(buffer),
    };
}

func (this *OngoingZipOperation) AddDir(path string, prefix string) error {
    if (this.writer == nil) {
        return fmt.Errorf("Can not add dir to closed zip operation: '%s'.", path);
    }

    return AddDirToZipWriter(path, prefix, this.writer);
}

func (this *OngoingZipOperation) AddFile(path string, archivePath string) error {
    if (this.writer == nil) {
        return fmt.Errorf("Can not add file to closed zip operation: '%s'.", path);
    }

    return AddFileToZipWriter(path, archivePath, this.writer);
}

func (this *OngoingZipOperation) GetBytes() []byte {
    this.Close();
    return this.buffer.Bytes();
}

func (this *OngoingZipOperation) Close() {
    if (this.writer == nil) {
        return;
    }

    this.writer.Close();
    this.writer = nil;
}

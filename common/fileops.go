package common

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

// Copy over assignment filespecs.
// 1) Do pre-copy operations.
// 2) Copy.
// 3) Do post-copy operations.
func CopyFileSpecs(
        sourceDir string, destDir string, opDir string,
        filespecs []FileSpec, onlyContents bool,
        preOps [][]string, postOps [][]string) error {
    var err error;

    // Do pre ops.
    for _, fileOperation := range preOps {
        err = doFileOperation(fileOperation, opDir);
        if (err != nil) {
            return fmt.Errorf("Failed to do pre file operation '%v': '%w'.", fileOperation, err);
        }
    }

    // Copy files.
    for _, filespec := range filespecs {
        err = filespec.CopyTarget(sourceDir, destDir, onlyContents);
        if (err != nil) {
            return fmt.Errorf("Failed to handle filespec '%s': '%w'", filespec, err);
        }
    }

    // Do post ops.
    for _, fileOperation := range postOps {
        err = doFileOperation(fileOperation, opDir);
        if (err != nil) {
            return fmt.Errorf("Failed to do post file operation '%v': '%w'.", fileOperation, err);
        }
    }

    return nil;
}

func doFileOperation(fileOperation []string, opDir string) error {
    if ((fileOperation == nil) || (len(fileOperation) == 0)) {
        return fmt.Errorf("File operation is empty.");
    }

    if (fileOperation[0] == "cp") {
        if (len(fileOperation) != 3) {
            return fmt.Errorf("Incorrect number of argument for 'cp' file operation. Expected 2, found %d.", len(fileOperation) - 1);
        }

        sourcePath := filepath.Join(opDir, fileOperation[1]);
        destPath := filepath.Join(opDir, fileOperation[2]);

        return util.CopyDirent(sourcePath, destPath, false);
    } else if (fileOperation[0] == "mv") {
        if (len(fileOperation) != 3) {
            return fmt.Errorf("Incorrect number of argument for 'mv' file operation. Expected 2, found %d.", len(fileOperation) - 1);
        }

        sourcePath := filepath.Join(opDir, fileOperation[1]);
        destPath := filepath.Join(opDir, fileOperation[2]);

        return os.Rename(sourcePath, destPath);
    } else {
        return fmt.Errorf("Unknown file operation: '%s'.", fileOperation[0]);
    }
}

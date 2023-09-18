package grader

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/eriq-augustine/autograder/util"
)

// Create a temp dir for grading as well as the three standard directories in it.
// Paths to the three direcotries (temp, in, out, work) will be returned.
// The created directory will be in the system's temp directory.
func prepTempGradingDir() (string, string, string, string, error) {
    tempDir, err := os.MkdirTemp("", "autograding-nodocker-");
    if (err != nil) {
        return "", "", "", "", fmt.Errorf("Could not create temp dir: '%w'.", err);
    }

    inputDir, outputDir, workDir, err := createStandardGradingDirs(tempDir);

    return tempDir, inputDir, outputDir, workDir, err;
}

// Create the standard three grading directories.
func createStandardGradingDirs(dir string) (string, string, string, error) {
    inputDir := filepath.Join(dir, GRADING_INPUT_DIRNAME);
    outputDir := filepath.Join(dir, GRADING_OUTPUT_DIRNAME);
    workDir := filepath.Join(dir, GRADING_WORK_DIRNAME);

    var err error;

    err = errors.Join(err, os.Mkdir(inputDir, 0755));
    err = errors.Join(err, os.Mkdir(outputDir, 0755));
    err = errors.Join(err, os.Mkdir(workDir, 0755));

    if (err != nil) {
        return "", "", "", fmt.Errorf("Could not create standard grading directories in temp dir ('%s'): '%w'.", dir, err);
    }

    return inputDir, outputDir, workDir, nil;
}

// Copy over assignment files.
// 1) Do pre-copy operations.
// 2) Copy.
// 3) Do post-copy operations.
func copyAssignmentFiles(sourceDir string, destDir string, opDir string,
                         files []string, onlyContents bool,
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
    for _, filename := range files {
        sourcePath := filepath.Join(sourceDir, filename);
        destPath := filepath.Join(destDir, filepath.Base(filename));

        if (onlyContents) {
            util.CopyDirContents(sourcePath, destPath);
        } else {
            util.CopyDirent(sourcePath, destPath, false);
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

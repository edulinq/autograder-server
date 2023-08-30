package grader

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const PYTHON_AUTOGRADER_INVOCATION = "python3 -m autograder.cli.grade-submission --grader <grader> --inputdir <inputdir> --outputdir <outputdir> --workdir <workdir> --outpath <outpath>"
const PYTHON_GRADER_FILENAME = "grader.py"

func RunNoDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions) (*model.GradedAssignment, error) {
    tempDir, inputDir, _, workDir, err := prepTempGradingDir();
    if (err != nil) {
        return nil, err;
    }

    if (!options.LeaveTempDir) {
        defer os.RemoveAll(tempDir);
    } else {
        log.Info().Str("tempdir", tempDir).Msg("Leaving behind temp grading dir.");
    }

    cmd, err := getAssignmentInvocation(assignment, inputDir, outputDir, workDir);
    if (err != nil) {
        return nil, err;
    }

    // Copy over the static files (and do any file ops).
    err = copyAssignmentFiles(filepath.Dir(assignment.SourcePath), workDir, tempDir,
            assignment.StaticFiles, false, assignment.PreStaticFileOperations, assignment.PostStaticFileOperations);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy static assignment files: '%w'.", err);
    }

    // Copy over the submission files (and do any file ops).
    err = copyAssignmentFiles(submissionPath, inputDir, tempDir,
            []string{"."}, true, assignment.PreSubmissionFileOperations, assignment.PostSubmissionFileOperations);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy submission ssignment files: '%w'.", err);
    }

    output, err := cmd.CombinedOutput();
    if (err != nil) {
        log.Warn().Str("assignment", assignment.FullID()).Str("tempdir", tempDir).Msg(string(output[:]));
        return nil, fmt.Errorf("Failed to run non-docker grader for assignment '%s': '%w'.", assignment.FullID(), err);
    }

    resultPath := filepath.Join(outputDir, model.GRADER_OUTPUT_RESULT_FILENAME);
    if (!util.PathExists(resultPath)) {
        return nil, fmt.Errorf("Cannot find output file ('%s') after non-docker grading.", resultPath);
    }

    var result model.GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, err;
    }

    return &result, nil;
}

// Get a command to invoke the non-docker grader.
func getAssignmentInvocation(assignment *model.Assignment, inputDir string, outputDir string, workDir string) (*exec.Cmd, error) {
    var rawCommand []string = nil;

    if ((assignment.Invocation != nil) && (len(assignment.Invocation) > 0)) {
        rawCommand = assignment.Invocation;
    }

    // Special case for the python grader (we know how to invoke that).
    if (assignment.Image == "autograder.python") {
        rawCommand = strings.Split(PYTHON_AUTOGRADER_INVOCATION, " ");
    }

    if (rawCommand == nil) {
        return nil, fmt.Errorf("Cannot get non-docker grader invocation for assignment: '%s'.", assignment.FullID());
    }

    cleanCommand := make([]string, 0, len(rawCommand));
    for _, value := range rawCommand {
        if (value == "<grader>") {
            value = filepath.Join(workDir, PYTHON_GRADER_FILENAME);
        } else if (value == "<inputdir>") {
            value = inputDir;
        } else if (value == "<outputdir>") {
            value = outputDir;
        } else if (value == "<workdir>") {
            value = workDir;
        } else if (value == "<outpath>") {
            value = filepath.Join(outputDir, model.GRADER_OUTPUT_RESULT_FILENAME);
        }

        cleanCommand = append(cleanCommand, value);
    }

    return exec.Command(cleanCommand[0], cleanCommand[1:]...), nil;
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

// Create a temp dir for grading as well as the three standard directories in it.
// Paths to the three direcotries (temp, in, out, work) will be returned.
// The created directory will be in the system's temp directory.
func prepTempGradingDir() (string, string, string, string, error) {
    tempDir, err := os.MkdirTemp("", "autograding-nodocker-");
    if (err != nil) {
        return "", "", "", "", fmt.Errorf("Could not create temp dir: '%w'.", err);
    }

    // Create the standard three grading directories.
    inputDir := filepath.Join(tempDir, GRADING_INPUT_DIRNAME);
    outputDir := filepath.Join(tempDir, GRADING_OUTPUT_DIRNAME);
    workDir := filepath.Join(tempDir, GRADING_WORK_DIRNAME);

    err = errors.Join(err, os.Mkdir(inputDir, 0755));
    err = errors.Join(err, os.Mkdir(outputDir, 0755));
    err = errors.Join(err, os.Mkdir(workDir, 0755));

    if (err != nil) {
        return "", "", "", "", fmt.Errorf("Could not create standard grading directories in temp dir ('%s'): '%w'.", tempDir, err);
    }

    return tempDir, inputDir, outputDir, workDir, nil;
}

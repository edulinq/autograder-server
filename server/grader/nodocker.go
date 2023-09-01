package grader

import (
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

func RunNoDockerGrader(assignment *model.Assignment, submissionPath string, outputDir string, options GradeOptions, gradingID string) (*model.GradedAssignment, error) {
    tempDir, inputDir, _, workDir, err := prepTempGradingDir();
    if (err != nil) {
        return nil, err;
    }

    if (!options.LeaveTempDir) {
        defer os.RemoveAll(tempDir);
    } else {
        log.Info().Str("path", tempDir).Msg("Leaving behind temp grading dir.");
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
            []string{"."}, true, [][]string{}, assignment.PostSubmissionFileOperations);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy submission ssignment files: '%w'.", err);
    }

    output, err := cmd.CombinedOutput();
    if (err != nil) {
        log.Warn().Str("assignment", assignment.FullID()).Str("tempdir", tempDir).Msg(string(output[:]));
        return nil, fmt.Errorf("Failed to run non-docker grader for assignment '%s': '%w'.", assignment.FullID(), err);
    }

    resultPath := filepath.Join(outputDir, GRADER_OUTPUT_RESULT_FILENAME);
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
            value = filepath.Join(outputDir, GRADER_OUTPUT_RESULT_FILENAME);
        }

        cleanCommand = append(cleanCommand, value);
    }

    return exec.Command(cleanCommand[0], cleanCommand[1:]...), nil;
}

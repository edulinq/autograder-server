package types

import (
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/util"
)

const (
    SUBMISSIONS_DIRNAME = "submissions"
    SUBMISSION_RESULT_FILENAME = "submission-result.json"
)

// Load submissions that are adjacent to a course config (if they exist).
func loadStaticSubmissions(courseConfigPath string) ([]*artifact.GradingResult, error) {
    submissions := make([]*artifact.GradingResult, 0);

    baseDir := util.ShouldAbs(filepath.Join(filepath.Dir(courseConfigPath), SUBMISSIONS_DIRNAME));
    if (!util.PathExists(baseDir)) {
        return submissions, nil;
    }

    resultPaths, err := util.FindFiles(SUBMISSION_RESULT_FILENAME, baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for submission results in '%s': '%w'.", baseDir, err);
    }

    for _, resultPath := range resultPaths {
        baseSubmissionDir := filepath.Dir(resultPath);
        submissionInputDir := filepath.Join(baseSubmissionDir, common.GRADING_INPUT_DIRNAME);

        var submissionResult artifact.GradedAssignment;
        err = util.JSONFromFile(resultPath, &submissionResult);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load submission result '%s': '%w'.", resultPath, err);
        }

        if (!util.PathExists(submissionInputDir)) {
            log.Warn().Str("dir", submissionInputDir).Msg("Input dir for submission result does not exist.");
            continue;
        }

        inputBytes, err := util.ZipToBytes(submissionInputDir, "", true);
        if (err != nil) {
            return nil, fmt.Errorf("Could not zip submission input dir '%s': '%w'.", submissionInputDir, err);
        }

        submissions = append(submissions, &artifact.GradingResult{
            Result: &submissionResult,
            InputFilesZip: inputBytes,
        });
    }

    return submissions, nil;
}

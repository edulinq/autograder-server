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
        gradingResult, err := LoadGradingResult(resultPath);
        if (err != nil) {
            return nil, err;
        }

        submissions = append(submissions, gradingResult);
    }

    return submissions, nil;
}

// Load a full standard grading result froma result path.
func LoadGradingResult(resultPath string) (*artifact.GradingResult, error) {
    baseSubmissionDir := filepath.Dir(resultPath);
    submissionInputDir := filepath.Join(baseSubmissionDir, common.GRADING_INPUT_DIRNAME);

    var gradingInfo artifact.GradingInfo;
    err := util.JSONFromFile(resultPath, &gradingInfo);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to load grading info '%s': '%w'.", resultPath, err);
    }

    if (!util.PathExists(submissionInputDir)) {
        return nil, fmt.Errorf("Input dir for submission result does not exist '%s': '%w'.", submissionInputDir, err);
    }

    fileContents, err := util.GzipDirectoryToBytes(submissionInputDir);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to gzip files in submission input dir '%s': '%w'.", submissionInputDir, err);
    }

    return &artifact.GradingResult{
        Info: &gradingInfo,
        InputFilesGZip: fileContents,
    }, nil;
}

func MustLoadGradingResult(resultPath string) *artifact.GradingResult {
    result, err := LoadGradingResult(resultPath);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", resultPath).Msg("Failed to load grading result.");
    }

    return result;
}

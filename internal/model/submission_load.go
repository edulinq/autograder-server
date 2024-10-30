package model

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	SUBMISSIONS_DIRNAME        = "submissions"
	SUBMISSION_RESULT_FILENAME = "submission-result.json"
)

// Load submissions that are adjacent to a course config (if they exist).
func loadStaticSubmissions(courseConfigPath string) ([]*GradingResult, error) {
	submissions := make([]*GradingResult, 0)

	baseDir := util.ShouldAbs(filepath.Join(filepath.Dir(courseConfigPath), SUBMISSIONS_DIRNAME))
	if !util.PathExists(baseDir) {
		return submissions, nil
	}

	resultPaths, err := util.FindFiles(SUBMISSION_RESULT_FILENAME, baseDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to search for submission results in '%s': '%w'.", baseDir, err)
	}

	for _, resultPath := range resultPaths {
		gradingResult, err := LoadGradingResult(resultPath)
		if err != nil {
			return nil, err
		}

		submissions = append(submissions, gradingResult)
	}

	return submissions, nil
}

// Load a full standard grading result from a result path.
func LoadGradingResult(resultPath string) (*GradingResult, error) {
	baseSubmissionDir := filepath.Dir(resultPath)
	submissionInputDir := filepath.Join(baseSubmissionDir, common.GRADING_INPUT_DIRNAME)
	submissionOutputDir := filepath.Join(baseSubmissionDir, common.GRADING_OUTPUT_DIRNAME)
	stdoutPath := filepath.Join(baseSubmissionDir, common.SUBMISSION_STDOUT_FILENAME)
	stderrPath := filepath.Join(baseSubmissionDir, common.SUBMISSION_STDERR_FILENAME)

	var gradingInfo GradingInfo
	err := util.JSONFromFile(resultPath, &gradingInfo)
	if err != nil {
		return nil, fmt.Errorf("Failed to load grading info '%s': '%w'.", resultPath, err)
	}

	if !util.PathExists(submissionInputDir) {
		return nil, fmt.Errorf("Input dir for submission result does not exist '%s': '%w'.", submissionInputDir, err)
	}

	if !util.PathExists(submissionOutputDir) {
		return nil, fmt.Errorf("Output dir for submission result does not exist '%s': '%w'.", submissionOutputDir, err)
	}

	inputFileContents, err := util.GzipDirectoryToBytes(submissionInputDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to gzip files in submission input dir '%s': '%w'.", submissionInputDir, err)
	}

	outputFileContents, err := util.GzipDirectoryToBytes(submissionOutputDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to gzip files in submission output dir '%s': '%w'.", submissionOutputDir, err)
	}

	stdout := ""
	if util.PathExists(stdoutPath) {
		stdout, err = util.ReadFile(stdoutPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read stdout file '%s': '%w'.", stdoutPath, err)
		}
	}

	stderr := ""
	if util.PathExists(stderrPath) {
		stderr, err = util.ReadFile(stderrPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read stderr file '%s': '%w'.", stderrPath, err)
		}
	}

	return &GradingResult{
		Info:            &gradingInfo,
		InputFilesGZip:  inputFileContents,
		OutputFilesGZip: outputFileContents,
		Stdout:          stdout,
		Stderr:          stderr,
	}, nil
}

func MustLoadGradingResult(resultPath string) *GradingResult {
	result, err := LoadGradingResult(resultPath)
	if err != nil {
		log.Fatal("Failed to load grading result.", err, log.NewAttr("path", resultPath))
	}

	return result
}

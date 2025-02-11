package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

func CreateFullSubmissionID(courseID string, assignmentID string, email string, shortSubmissionID string) string {
	return util.JoinStrings(SUBMISSION_ID_DELIM, courseID, assignmentID, email, shortSubmissionID)
}

// Split a full submission ID and return: course id, assignment id, user email, and short id.
func SplitFullSubmissionID(fullSubmissionID string) (string, string, string, string, error) {
	parts := splitFullSubmissionID(fullSubmissionID)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("Malformed full submission ID '%s'. Expected 4 components, found %d.", fullSubmissionID, len(parts))
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func splitFullSubmissionID(fullSubmissionID string) []string {
	return strings.Split(fullSubmissionID, SUBMISSION_ID_DELIM)
}

// Get the short submission ID from either a full or short submission ID.
// Accepted inputs are: full id, short id, empty string.
func GetShortSubmissionID(submissionID string) string {
	parts := splitFullSubmissionID(submissionID)
	return parts[len(parts)-1]
}

// Create a temp dir for grading as well as the three standard directories in it.
// Paths to the three direcotries (temp, in, out, work) will be returned.
// The created directory will be in the system's temp directory.
func PrepTempGradingDir(name string) (string, string, string, string, error) {
	tempDir, err := util.MkDirTemp(fmt.Sprintf("autograder-%s-", name))
	if err != nil {
		return "", "", "", "", fmt.Errorf("Could not create temp dir: '%w'.", err)
	}

	inputDir, outputDir, workDir, err := CreateStandardGradingDirs(tempDir)

	return tempDir, inputDir, outputDir, workDir, err
}

// Create the standard three grading directories.
func CreateStandardGradingDirs(dir string) (string, string, string, error) {
	inputDir := filepath.Join(dir, GRADING_INPUT_DIRNAME)
	outputDir := filepath.Join(dir, GRADING_OUTPUT_DIRNAME)
	workDir := filepath.Join(dir, GRADING_WORK_DIRNAME)

	var err error

	err = errors.Join(err, os.Mkdir(inputDir, 0755))
	err = errors.Join(err, os.Mkdir(outputDir, 0755))
	err = errors.Join(err, os.Mkdir(workDir, 0755))

	if err != nil {
		return "", "", "", fmt.Errorf("Could not create standard grading directories in temp dir ('%s'): '%w'.", dir, err)
	}

	return inputDir, outputDir, workDir, nil
}

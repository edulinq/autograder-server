package grader

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type TestSubmissionInfo struct {
	ID             string
	Dir            string
	Files          []string
	TestSubmission *model.TestSubmission
	Assignment     *model.Assignment
}

func GetTestSubmissions(baseDir string, useDocker bool) ([]*TestSubmissionInfo, error) {
	testSubmissionPaths, err := util.FindFiles("test-submission.json", baseDir)
	if err != nil {
		return nil, fmt.Errorf("Could not find test results in '%s': '%w'.", baseDir, err)
	}

	testSubmissions := make([]*TestSubmissionInfo, 0, len(testSubmissionPaths))

	for _, testSubmissionPath := range testSubmissionPaths {
		testSubmissionPath = util.ShouldAbs(testSubmissionPath)

		var testSubmission model.TestSubmission
		err := util.JSONFromFile(testSubmissionPath, &testSubmission)
		if err != nil {
			return nil, fmt.Errorf("Failed to load test submission: '%s': '%w'.", testSubmissionPath, err)
		}

		assignment, err := fetchTestSubmissionAssignment(testSubmissionPath)
		if err != nil {
			return nil, fmt.Errorf("Could not find assignment for test submission '%s': '%w'.", testSubmissionPath, err)
		}

		// Skip non-bash test assignments when not using Docker.
		if !useDocker && assignment.ID != "bash" {
			continue
		}

		dir := util.ShouldAbs(filepath.Dir(testSubmissionPath))

		paths, err := util.GetAllDirents(dir, false, false)
		if err != nil {
			return nil, fmt.Errorf("Failed to get test submission files: '%w'.", err)
		}

		removeIndex := slices.Index(paths, testSubmissionPath)
		paths = slices.Delete(paths, removeIndex, removeIndex+1)

		testSubmissions = append(testSubmissions, &TestSubmissionInfo{
			ID:             strings.TrimPrefix(testSubmissionPath, baseDir),
			Dir:            dir,
			Files:          paths,
			TestSubmission: &testSubmission,
			Assignment:     assignment,
		})
	}

	if len(testSubmissions) == 0 {
		return nil, fmt.Errorf("Could not find any test submissions in '%s'.", baseDir)
	}

	return testSubmissions, nil
}

func CheckAndClearIDs(test *testing.T, i int, expectedResults map[string]*model.SubmissionHistoryItem, actualResults map[string]*model.SubmissionHistoryItem) bool {
	for user, expected := range expectedResults {
		actual, ok := actualResults[user]
		if !ok {
			test.Errorf("Case %d: Unable to find result for user '%s'. Expected: '%v'.",
				i, user, util.MustToJSONIndent(expected))
			return true
		}

		if (expected == nil) && (actual == nil) {
			return false
		}

		if expected == nil {
			test.Errorf("Case %d: Unexpected result for user '%s'. Expected: '<nil>', actual: '%s'.",
				i, user, util.MustToJSONIndent(actual))
			return true
		}

		if actual == nil {
			test.Errorf("Case %d: Unexpected result for user '%s'. Expected: '%s', actual: '<nil>'.",
				i, user, util.MustToJSONIndent(expected))
			return true
		}

		if expected.ShortID == actual.ShortID {
			test.Errorf("Case %d: Regrade submission has the same short ID as the previous submission: '%s'.", i, expected.ShortID)
			return true
		}

		if expected.ID == actual.ID {
			test.Errorf("Case %d: Regrade submission has the same ID as the previous submission: '%s'.", i, expected.ID)
			return true
		}

		actual.ShortID = ""
		expected.ShortID = ""
		actual.ID = ""
		expected.ID = ""
	}

	return false
}

// Test submission are within their assignment's directory,
// just check the source dirs for existing courses and assignments.
func fetchTestSubmissionAssignment(testSubmissionPath string) (*model.Assignment, error) {
	testSubmissionPath = util.ShouldAbs(testSubmissionPath)

	assignmentPath := util.SearchParents(testSubmissionPath, model.ASSIGNMENT_CONFIG_FILENAME)
	if assignmentPath == "" {
		return nil, fmt.Errorf("Could not find assignment file for test submission.")
	}

	coursePath := util.SearchParents(testSubmissionPath, model.COURSE_CONFIG_FILENAME)
	if coursePath == "" {
		return nil, fmt.Errorf("Could not find course file for test submission.")
	}

	course, err := model.ReadCourseConfig(coursePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load course '%s': '%w'.", coursePath, err)
	}

	relDir, _ := filepath.Rel(filepath.Dir(util.ShouldAbs(coursePath)), filepath.Dir(util.ShouldAbs(assignmentPath)))
	assignment, err := model.ReadAssignmentConfig(course, assignmentPath, relDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to load assignment '%s': '%w'.", assignmentPath, err)
	}

	return db.GetAssignment(course.GetID(), assignment.GetID())
}

func getTestSubmissionResultPath(shortID string) string {
	return filepath.Join(config.GetTestdataDir(), "course101", "submissions", "HW0", "course-student@test.edulinq.org", shortID, "submission-result.json")
}

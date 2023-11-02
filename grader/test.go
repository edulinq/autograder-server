package grader

import (
    "fmt"
    "path/filepath"
    "slices"
    "strings"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type TestSubmissionInfo struct {
    ID string
    Dir string
    Files []string
    TestSubmission *artifact.TestSubmission
    Assignment *model.Assignment

}

func GetTestSubmissions(baseDir string) ([]*TestSubmissionInfo, error) {
    err := LoadCourses()
    if (err != nil) {
        return nil, fmt.Errorf("Could not load courses: '%w'.", err);
    }

    testSubmissionPaths, err := util.FindFiles("test-submission.json", baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Could not find test results in '%s': '%w'.", baseDir, err);
    }

    testSubmissions := make([]*TestSubmissionInfo, 0, len(testSubmissionPaths));

    for _, testSubmissionPath := range testSubmissionPaths {
        testSubmissionPath = util.ShouldAbs(testSubmissionPath);

        var testSubmission artifact.TestSubmission;
        err := util.JSONFromFile(testSubmissionPath, &testSubmission);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load test submission: '%s': '%w'.", testSubmissionPath, err);
        }

        assignment := fetchTestSubmissionAssignment(testSubmissionPath);
        if (assignment == nil) {
            return nil, fmt.Errorf("Could not find assignment for test submission '%s'.", testSubmissionPath);
        }

        dir := util.ShouldAbs(filepath.Dir(testSubmissionPath));

        paths, err := util.GetAllDirents(dir);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to get test submission files: '%w'.", err);
        }

        removeIndex := slices.Index(paths, testSubmissionPath);
        paths = slices.Delete(paths, removeIndex, removeIndex + 1);

        testSubmissions = append(testSubmissions, &TestSubmissionInfo{
            ID: strings.TrimPrefix(testSubmissionPath, baseDir),
            Dir: dir,
            Files: paths,
            TestSubmission: &testSubmission,
            Assignment: assignment,
        });
    }

    return testSubmissions, nil;
}

// Test submission are withing their assignment's directory,
// just check the source dirs for existing courses and assignments.
func fetchTestSubmissionAssignment(testSubmissionPath string) *model.Assignment {
    testSubmissionPath = util.ShouldAbs(testSubmissionPath);

    for _, course := range GetCourses() {
        if (!util.PathHasParent(testSubmissionPath, course.GetSourceDir())) {
            continue;
        }

        for _, assignment := range course.GetAssignments() {
            if (util.PathHasParent(testSubmissionPath, assignment.GetSourceDir())) {
                return assignment;
            }
        }
    }

    return nil;
}

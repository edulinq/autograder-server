package disk

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func (this *backend) saveSubmissionsLock(course *types.Course, submissions []*artifact.GradingResult, acquireLock bool) error {
    if (acquireLock) {
        this.lock.Lock();
        defer this.lock.Unlock();
    }

    for _, submission := range submissions {
        baseDir := this.getSubmissionDirFromResult(submission.Result);
        err := util.MkDir(baseDir);
        if (err != nil) {
            return fmt.Errorf("Failed to make submission dir '%s': '%w'.", baseDir, err);
        }

        resultPath := filepath.Join(baseDir, types.SUBMISSION_RESULT_FILENAME);
        err = util.ToJSONFileIndent(submission.Result, resultPath);
        if (err != nil) {
            return fmt.Errorf("Failed to write submission result '%s': '%w'.", resultPath, err);
        }

        err = util.UnzipFromBytes(submission.InputFilesZip, baseDir);
        if (err != nil) {
            return fmt.Errorf("Failed to input files into '%s': '%w'.", baseDir, err);
        }
    }

    return nil;
}

func (this *backend) SaveSubmissions(course *types.Course, submissions []*artifact.GradingResult) error {
    return this.saveSubmissionsLock(course, submissions, true);
}

func (this *backend) GetNextSubmissionID(assignment *types.Assignment, email string) (string, error) {
    submissionID := time.Now().Unix();
    baseDir := this.getUserSubmissionDir(assignment.Course.GetID(), assignment.GetID(), email);

    for ; ; {
        path := filepath.Join(baseDir, fmt.Sprintf("%d", submissionID));
        if (!util.PathExists(path)) {
            break;
        }

        // This ID has been used.
        submissionID++;
    }

    return fmt.Sprintf("%d", submissionID), nil;
}

func (this *backend) GetSubmissionResult(assignment *types.Assignment, email string, shortSubmissionID string) (*artifact.GradedAssignment, error) {
    var err error;

    if (shortSubmissionID == "") {
        shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to get most recent submission id: '%w'.", err);
        }
    }

    submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID);
    resultPath := filepath.Join(submissionDir, types.SUBMISSION_RESULT_FILENAME);

    if (!util.PathExists(resultPath)) {
        return nil, nil;
    }

    var result artifact.GradedAssignment;
    err = util.JSONFromFile(resultPath, &result);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to deserialize submission result '%s': '%w'.", resultPath, err);
    }

    return &result, nil;
}

func (this *backend) GetSubmissionHistory(assignment *types.Assignment, email string) ([]*artifact.SubmissionHistoryItem, error) {
    history := make([]*artifact.SubmissionHistoryItem, 0);

    submissionsDir := this.getUserSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), email);
    if (!util.PathExists(submissionsDir)) {
        return history, nil;
    }

    dirents, err := os.ReadDir(submissionsDir);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to read user submissions dir '%s': '%w'.", submissionsDir, err);
    }

    if (len(dirents) == 0) {
        return history, nil;
    }

    for _, dirent := range dirents {
        resultPath := filepath.Join(submissionsDir, dirent.Name(), types.SUBMISSION_RESULT_FILENAME);

        var result artifact.GradedAssignment;
        err = util.JSONFromFile(resultPath, &result);
        if (err != nil) {
            return nil, fmt.Errorf("Unable to deserialize submission result '%s': '%w'.", resultPath, err);
        }

        history = append(history, result.ToHistoryItem());
    }

    return history, nil;
}

// TEST
func (this *backend) GetScoringInfos(assignment *types.Assignment, onlyRole usr.UserRole) (map[string]*artifact.ScoringInfo, error) {
    scoringInfos := make(map[string]*artifact.ScoringInfo);

    // TEST
    return scoringInfos, nil;
}

// TEST
func (this *backend) GetRecentSubmissions(assignment *types.Assignment, onlyRole usr.UserRole) (map[string]*artifact.GradedAssignment, error) {
    results := make(map[string]*artifact.GradedAssignment);

    // TEST
    return results, nil;
}

// TEST
func (this *backend) GetRecentSubmissionSurvey(assignment *types.Assignment, onlyRole usr.UserRole) (map[string]*artifact.SubmissionHistoryItem, error) {
    results := make(map[string]*artifact.SubmissionHistoryItem);

    // TEST
    return results, nil;
}

func (this *backend) getSubmissionDir(courseID string, assignmentID string, user string, shortSubmissionID string) string {
    return filepath.Join(this.getUserSubmissionDir(courseID, assignmentID, user), shortSubmissionID);
}

func (this *backend) getSubmissionDirFromAssignment(assignment *types.Assignment, user string, shortSubmissionID string) string {
    return this.getSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), user, shortSubmissionID);
}

func (this *backend) getSubmissionDirFromResult(submission *artifact.GradedAssignment) string {
    return this.getSubmissionDir(submission.CourseID, submission.AssignmentID, submission.User, submission.ShortID);
}

func (this *backend) getUserSubmissionDir(courseID string, assignmentID string, user string) string {
    return filepath.Join(this.getCourseDirFromID(courseID), types.SUBMISSIONS_DIRNAME, assignmentID, user);
}

// Get the short id of the most recent submission (or empty string if there are no submissions).
func (this *backend) getMostRecentSubmissionID(assignment *types.Assignment, email string) (string, error) {
    submissionsDir := this.getUserSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), email);
    if (!util.PathExists(submissionsDir)) {
        return "", nil;
    }

    dirents, err := os.ReadDir(submissionsDir);
    if (err != nil) {
        return "", fmt.Errorf("Unable to read user submissions dir '%s': '%w'.", submissionsDir, err);
    }

    if (len(dirents) == 0) {
        return "", nil;
    }

    return dirents[len(dirents) - 1].Name(), nil;
}

/* TEST
// Get all the paths to the submission files for an assignment and user.
// The results will be sorted in ascending order (first submission first).
// An empty slice indicates that there are no matching submission files.
func (this *Assignment) getSubmissionFiles(user string, filename string) ([]string, error) {
    submissionsDir, err := this.getSubmissionsDir();
    if (err != nil) {
        return nil, err;
    }

    paths := make([]string, 0);

    baseDir := filepath.Join(submissionsDir, user);
    if (!util.PathExists(baseDir)) {
        return paths, nil;
    }

    if (!util.IsDir(baseDir)) {
        return nil, fmt.Errorf("Expected user's submission dir '%s' exists and is not a dir.", baseDir);
    }

    dirents, err := os.ReadDir(baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to read dir '%s': '%w'.", baseDir, err);
    }

    for _, dirent := range dirents {
        if (!dirent.IsDir()) {
            continue;
        }

        path := filepath.Join(baseDir, dirent.Name(), common.GRADING_OUTPUT_DIRNAME, filename);
        if (!util.IsFile(path)) {
            continue;
        }

        paths = append(paths, path);
    }

    return paths, nil;
}

// Get all the paths to the most recent submission file for each user for this assignment.
// The returned map will contain an entry for every user (if not nil).
// An empty entry in the map indicates the user has no submissions.
func (this *Assignment) getAllRecentSubmissionFiles(users map[string]*usr.User, filename string) (map[string]string, error) {
    paths := make(map[string]string);

    for email, _ := range users {
        userPaths, err := this.getSubmissionFiles(email, filename);
        if (err != nil) {
            return nil, err;
        }

        if (len(userPaths) == 0) {
            paths[email] = "";
        } else {
            paths[email] = userPaths[len(userPaths) - 1];
        }
    }

    return paths, nil;
}
*/

package disk

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func (this *backend) saveSubmissionsLock(course *model.Course, submissions []*artifact.GradingResult, acquireLock bool) error {
    if (acquireLock) {
        this.lock.Lock();
        defer this.lock.Unlock();
    }

    for _, submission := range submissions {
        baseDir := this.getSubmissionDirFromResult(submission.Info);
        err := util.MkDir(baseDir);
        if (err != nil) {
            return fmt.Errorf("Failed to make submission dir '%s': '%w'.", baseDir, err);
        }

        resultPath := filepath.Join(baseDir, model.SUBMISSION_RESULT_FILENAME);
        err = util.ToJSONFileIndent(submission.Info, resultPath);
        if (err != nil) {
            return fmt.Errorf("Failed to write submission result '%s': '%w'.", resultPath, err);
        }

        err = util.GzipBytesToDirectory(filepath.Join(baseDir, common.GRADING_INPUT_DIRNAME), submission.InputFilesGZip);
        if (err != nil) {
            return fmt.Errorf("Failed to write submission input files: '%w'.", err);
        }

        err = util.GzipBytesToDirectory(filepath.Join(baseDir, common.GRADING_OUTPUT_DIRNAME), submission.OutputFilesGZip);
        if (err != nil) {
            return fmt.Errorf("Failed to write submission input files: '%w'.", err);
        }

        err = util.WriteFile(submission.Stdout, filepath.Join(baseDir, common.SUBMISSION_STDOUT_FILENAME));
        if (err != nil) {
            return fmt.Errorf("Failed to write submission stdout file: '%w'.", err);
        }

        err = util.WriteFile(submission.Stderr, filepath.Join(baseDir, common.SUBMISSION_STDERR_FILENAME));
        if (err != nil) {
            return fmt.Errorf("Failed to write submission stderr file: '%w'.", err);
        }
    }

    return nil;
}

func (this *backend) SaveSubmissions(course *model.Course, submissions []*artifact.GradingResult) error {
    return this.saveSubmissionsLock(course, submissions, true);
}

func (this *backend) GetNextSubmissionID(assignment *model.Assignment, email string) (string, error) {
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

func (this *backend) GetSubmissionResult(assignment *model.Assignment, email string, shortSubmissionID string) (*artifact.GradingInfo, error) {
    var err error;

    if (shortSubmissionID == "") {
        shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to get most recent submission id: '%w'.", err);
        }
    }

    if (shortSubmissionID == "") {
        return nil, nil;
    }

    submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID);
    resultPath := filepath.Join(submissionDir, model.SUBMISSION_RESULT_FILENAME);

    if (!util.PathExists(resultPath)) {
        return nil, nil;
    }

    var gradingInfo artifact.GradingInfo;
    err = util.JSONFromFile(resultPath, &gradingInfo);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err);
    }

    return &gradingInfo, nil;
}

func (this *backend) GetSubmissionHistory(assignment *model.Assignment, email string) ([]*artifact.SubmissionHistoryItem, error) {
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
        resultPath := filepath.Join(submissionsDir, dirent.Name(), model.SUBMISSION_RESULT_FILENAME);

        var gradingInfo artifact.GradingInfo;
        err = util.JSONFromFile(resultPath, &gradingInfo);
        if (err != nil) {
            return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err);
        }

        history = append(history, gradingInfo.ToHistoryItem());
    }

    return history, nil;
}

func (this *backend) GetRecentSubmissions(assignment *model.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingInfo, error) {
    gradingInfos := make(map[string]*artifact.GradingInfo);

    users, err := this.GetUsers(assignment.Course);
    if (err != nil) {
        return nil, err;
    }

    for email, user := range users {
        if ((filterRole != usr.Unknown) && (filterRole != user.Role)) {
            continue;
        }

        shortSubmissionID, err := this.getMostRecentSubmissionID(assignment, email);
        if (err != nil) {
            return nil, err;
        }

        if (shortSubmissionID == "") {
            gradingInfos[email] = nil;
            continue;
        }

        resultPath := filepath.Join(this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID), model.SUBMISSION_RESULT_FILENAME);

        var gradingInfo artifact.GradingInfo;
        err = util.JSONFromFile(resultPath, &gradingInfo);
        if (err != nil) {
            return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err);
        }

        gradingInfos[email] = &gradingInfo;
    }

    return gradingInfos, nil;
}

func (this *backend) GetScoringInfos(assignment *model.Assignment, filterRole usr.UserRole) (map[string]*artifact.ScoringInfo, error) {
    scoringInfos := make(map[string]*artifact.ScoringInfo);

    submissionResults, err := this.GetRecentSubmissions(assignment, filterRole);
    if (err != nil) {
        return nil, err;
    }

    for email, submissionResult := range submissionResults {
        if (submissionResult == nil) {
            scoringInfos[email] = nil;
        } else {
            scoringInfos[email] = submissionResult.ToScoringInfo();
        }
    }

    return scoringInfos, nil;
}

func (this *backend) GetRecentSubmissionSurvey(assignment *model.Assignment, filterRole usr.UserRole) (map[string]*artifact.SubmissionHistoryItem, error) {
    results := make(map[string]*artifact.SubmissionHistoryItem);

    submissionResults, err := this.GetRecentSubmissions(assignment, filterRole);
    if (err != nil) {
        return nil, err;
    }

    for email, submissionResult := range submissionResults {
        if (submissionResult == nil) {
            results[email] = nil;
        } else {
            results[email] = submissionResult.ToHistoryItem();
        }
    }

    return results, nil;
}

func (this *backend) GetSubmissionContents(assignment *model.Assignment, email string, shortSubmissionID string) (*artifact.GradingResult, error) {
    var err error;

    if (shortSubmissionID == "") {
        shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to get most recent submission id: '%w'.", err);
        }
    }

    if (shortSubmissionID == "") {
        return nil, nil;
    }

    submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID);
    resultPath := filepath.Join(submissionDir, model.SUBMISSION_RESULT_FILENAME);

    if (!util.PathExists(resultPath)) {
        return nil, nil;
    }

    return model.LoadGradingResult(resultPath);
}

func (this *backend) GetRecentSubmissionContents(assignment *model.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingResult, error) {
    results := make(map[string]*artifact.GradingResult);

    users, err := this.GetUsers(assignment.Course);
    if (err != nil) {
        return nil, err;
    }

    for email, user := range users {
        if ((filterRole != usr.Unknown) && (filterRole != user.Role)) {
            continue;
        }

        result, err := this.GetSubmissionContents(assignment, email, "");
        if (err != nil) {
            return nil, err;
        }

        results[email] = result;
    }

    return results, nil;
}

func (this *backend) getSubmissionDir(courseID string, assignmentID string, user string, shortSubmissionID string) string {
    return filepath.Join(this.getUserSubmissionDir(courseID, assignmentID, user), shortSubmissionID);
}

func (this *backend) getSubmissionDirFromAssignment(assignment *model.Assignment, user string, shortSubmissionID string) string {
    return this.getSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), user, shortSubmissionID);
}

func (this *backend) getSubmissionDirFromResult(gradingInfo *artifact.GradingInfo) string {
    return this.getSubmissionDir(gradingInfo.CourseID, gradingInfo.AssignmentID, gradingInfo.User, gradingInfo.ShortID);
}

func (this *backend) getUserSubmissionDir(courseID string, assignmentID string, user string) string {
    return filepath.Join(this.getCourseDirFromID(courseID), model.SUBMISSIONS_DIRNAME, assignmentID, user);
}

// Get the short id of the most recent submission (or empty string if there are no submissions).
func (this *backend) getMostRecentSubmissionID(assignment *model.Assignment, email string) (string, error) {
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

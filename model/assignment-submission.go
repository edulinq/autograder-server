package model

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func (this *Assignment) getSubmissionsDir() (string, error) {
    assignmentDir := filepath.Dir(this.SourcePath);
    path := filepath.Join(assignmentDir, DEFAULT_SUBMISSIONS_DIR);

    if (util.PathExists(path)) {
        if (!util.IsDir(path)) {
            return "", fmt.Errorf("Submissions dir ('%s') already exists and is not a dir.", path);
        }
    } else {
        err := os.MkdirAll(path, 0755);
        if (err != nil) {
            return "", fmt.Errorf("Failed to make submissions directory ('%s'): '%w'.", path, err);
        }
    }

    return path, nil;
}

func (this *Assignment) PrepareSubmission(user string) (string, int64, error) {
    submissionsDir, err := this.getSubmissionsDir();
    if (err != nil) {
        return "", 0, err;
    }

    return this.PrepareSubmissionWithDir(user, submissionsDir);
}

// Prepare a place to hold the student's submission history.
func (this *Assignment) PrepareSubmissionWithDir(user string, submissionsDir string) (string, int64, error) {
    submissionID := time.Now().Unix();
    var path string;

    for ; ; {
        path = filepath.Join(submissionsDir, user, fmt.Sprintf("%d", submissionID));
        if (!util.PathExists(path)) {
            break;
        }

        // This ID has been used.
        submissionID++;
    }

    err := os.MkdirAll(path, 0755);
    if (err != nil) {
        return "", 0, fmt.Errorf("Failed to make submission directory ('%s'): '%w'.", path, err);
    }

    return path, submissionID, nil;
}

// See getSubmissionFiles().
// Fetches full grading result.
func (this *Assignment) GetSubmissionResults(user string) ([]string, error) {
    return this.getSubmissionFiles(user, common.GRADER_OUTPUT_RESULT_FILENAME);
}

// See getSubmissionFiles().
// Fetches grading summary.
func (this *Assignment) GetSubmissionSummaries(user string) ([]string, error) {
    return this.getSubmissionFiles(user, common.GRADER_OUTPUT_SUMMARY_FILENAME);
}

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

// See getAllRecentSubmissionFiles().
// Fetches full grading result.
func (this *Assignment) GetAllRecentSubmissionResults(users map[string]*usr.User) (map[string]string, error) {
    return this.getAllRecentSubmissionFiles(users, common.GRADER_OUTPUT_RESULT_FILENAME);
}

// See getAllRecentSubmissionFiles().
// Fetches grading summary.
func (this *Assignment) GetAllRecentSubmissionSummaries(users map[string]*usr.User) (map[string]string, error) {
    return this.getAllRecentSubmissionFiles(users, common.GRADER_OUTPUT_SUMMARY_FILENAME);
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

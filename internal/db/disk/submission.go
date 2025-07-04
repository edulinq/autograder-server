package disk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *backend) saveSubmissionsLock(course *model.Course, submissions []*model.GradingResult, acquireLock bool) error {
	var errs error = nil

	for _, submission := range submissions {
		errs = errors.Join(errs, this.saveSubmissionLock(course, submission, acquireLock))
	}

	return errs
}

func (this *backend) saveSubmissionLock(course *model.Course, submission *model.GradingResult, acquireLock bool) error {
	baseDir := this.getSubmissionDirFromResult(submission.Info)

	if acquireLock {
		this.contextLock(baseDir)
		defer this.contextUnlock(baseDir)
	}

	err := util.MkDir(baseDir)
	if err != nil {
		return fmt.Errorf("Failed to make submission dir '%s': '%w'.", baseDir, err)
	}

	resultPath := filepath.Join(baseDir, model.SUBMISSION_RESULT_FILENAME)
	err = util.ToJSONFileIndent(submission.Info, resultPath)
	if err != nil {
		return fmt.Errorf("Failed to write submission result '%s': '%w'.", resultPath, err)
	}

	err = util.GzipBytesToDirectory(filepath.Join(baseDir, common.GRADING_INPUT_DIRNAME), submission.InputFilesGZip)
	if err != nil {
		return fmt.Errorf("Failed to write submission input files: '%w'.", err)
	}

	err = util.GzipBytesToDirectory(filepath.Join(baseDir, common.GRADING_OUTPUT_DIRNAME), submission.OutputFilesGZip)
	if err != nil {
		return fmt.Errorf("Failed to write submission input files: '%w'.", err)
	}

	err = util.WriteFile(submission.Stdout, filepath.Join(baseDir, common.SUBMISSION_STDOUT_FILENAME))
	if err != nil {
		return fmt.Errorf("Failed to write submission stdout file: '%w'.", err)
	}

	err = util.WriteFile(submission.Stderr, filepath.Join(baseDir, common.SUBMISSION_STDERR_FILENAME))
	if err != nil {
		return fmt.Errorf("Failed to write submission stderr file: '%w'.", err)
	}

	return nil
}

func (this *backend) SaveSubmissions(course *model.Course, submissions []*model.GradingResult) error {
	return this.saveSubmissionsLock(course, submissions, true)
}

// Compute the next submission ID based on the current time.
// If a submission ID exists for the current time, increment the ID until it is unique.
// The next submission ID will be the largest ID for the user on this assignment.
func (this *backend) GetNextSubmissionID(assignment *model.Assignment, email string) (string, error) {
	submissionID := time.Now().Unix()
	baseDir := this.getUserSubmissionDir(assignment.Course.GetID(), assignment.GetID(), email)

	for {
		path := filepath.Join(baseDir, fmt.Sprintf("%d", submissionID))
		if !util.PathExists(path) {
			break
		}

		// This ID has been used.
		submissionID++
	}

	return fmt.Sprintf("%d", submissionID), nil
}

func (this *backend) GetPreviousSubmissionID(assignment *model.Assignment, email string, shortSubmissionID string) (string, error) {
	history, err := this.GetSubmissionHistory(assignment, email)
	if err != nil {
		return "", err
	}

	if len(history) <= 1 {
		return "", nil
	}

	index := -1
	for i, item := range history {
		if item.ShortID == shortSubmissionID {
			index = i
			break
		}
	}

	if index <= 0 {
		return "", nil
	}

	return history[index-1].ID, nil
}

func (this *backend) GetSubmissionResult(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingInfo, error) {
	var err error

	if shortSubmissionID == "" {
		shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email)
		if err != nil {
			return nil, fmt.Errorf("Failed to get most recent submission id: '%w'.", err)
		}
	}

	if shortSubmissionID == "" {
		return nil, nil
	}

	submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID)
	resultPath := filepath.Join(submissionDir, model.SUBMISSION_RESULT_FILENAME)

	if !util.PathExists(resultPath) {
		return nil, nil
	}

	var gradingInfo model.GradingInfo
	err = util.JSONFromFile(resultPath, &gradingInfo)
	if err != nil {
		return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err)
	}

	return &gradingInfo, nil
}

func (this *backend) GetSubmissionHistory(assignment *model.Assignment, email string) ([]*model.SubmissionHistoryItem, error) {
	history := make([]*model.SubmissionHistoryItem, 0)

	submissionsDir := this.getUserSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), email)
	if !util.PathExists(submissionsDir) {
		return history, nil
	}

	dirents, err := os.ReadDir(submissionsDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to read user submissions dir '%s': '%w'.", submissionsDir, err)
	}

	if len(dirents) == 0 {
		return history, nil
	}

	for _, dirent := range dirents {
		resultPath := filepath.Join(submissionsDir, dirent.Name(), model.SUBMISSION_RESULT_FILENAME)

		var gradingInfo model.GradingInfo
		err = util.JSONFromFile(resultPath, &gradingInfo)
		if err != nil {
			return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err)
		}

		history = append(history, gradingInfo.ToHistoryItem())
	}

	slices.SortFunc(history, func(a *model.SubmissionHistoryItem, b *model.SubmissionHistoryItem) int {
		return int((a.GradingStartTime - b.GradingStartTime).ToMSecs())
	})

	return history, nil
}

func (this *backend) GetRecentSubmissions(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.GradingInfo, error) {
	gradingInfos := make(map[string]*model.GradingInfo)

	users, err := this.GetCourseUsers(assignment.Course)
	if err != nil {
		return nil, err
	}

	for email, user := range users {
		if !reference.RefersTo(email, user.Role) {
			continue
		}

		shortSubmissionID, err := this.getMostRecentSubmissionID(assignment, email)
		if err != nil {
			return nil, err
		}

		if shortSubmissionID == "" {
			gradingInfos[email] = nil
			continue
		}

		resultPath := filepath.Join(this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID), model.SUBMISSION_RESULT_FILENAME)

		var gradingInfo model.GradingInfo
		err = util.JSONFromFile(resultPath, &gradingInfo)
		if err != nil {
			return nil, fmt.Errorf("Unable to deserialize grading info '%s': '%w'.", resultPath, err)
		}

		gradingInfos[email] = &gradingInfo
	}

	return gradingInfos, nil
}

func (this *backend) GetScoringInfos(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.ScoringInfo, error) {
	scoringInfos := make(map[string]*model.ScoringInfo)

	submissionResults, err := this.GetRecentSubmissions(assignment, reference)
	if err != nil {
		return nil, err
	}

	for email, submissionResult := range submissionResults {
		if submissionResult == nil {
			scoringInfos[email] = nil
		} else {
			scoringInfos[email] = submissionResult.ToScoringInfo()
		}
	}

	return scoringInfos, nil
}

func (this *backend) GetRecentSubmissionSurvey(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.SubmissionHistoryItem, error) {
	results := make(map[string]*model.SubmissionHistoryItem)

	submissionResults, err := this.GetRecentSubmissions(assignment, reference)
	if err != nil {
		return nil, err
	}

	for email, submissionResult := range submissionResults {
		if submissionResult == nil {
			results[email] = nil
		} else {
			results[email] = submissionResult.ToHistoryItem()
		}
	}

	return results, nil
}

func (this *backend) GetSubmissionContents(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingResult, error) {
	var err error

	if shortSubmissionID == "" {
		shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email)
		if err != nil {
			return nil, fmt.Errorf("Failed to get most recent submission id: '%w'.", err)
		}
	}

	if shortSubmissionID == "" {
		return nil, nil
	}

	submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID)
	resultPath := filepath.Join(submissionDir, model.SUBMISSION_RESULT_FILENAME)

	if !util.PathExists(resultPath) {
		return nil, nil
	}

	return model.LoadGradingResult(resultPath)
}

func (this *backend) GetRecentSubmissionContents(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.GradingResult, error) {
	results := make(map[string]*model.GradingResult)

	users, err := this.GetCourseUsers(assignment.Course)
	if err != nil {
		return nil, err
	}

	for email, user := range users {
		if !reference.RefersTo(email, user.Role) {
			continue
		}

		result, err := this.GetSubmissionContents(assignment, email, "")
		if err != nil {
			return nil, err
		}

		results[email] = result
	}

	return results, nil
}

func (this *backend) getSubmissionDir(courseID string, assignmentID string, user string, shortSubmissionID string) string {
	return filepath.Join(this.getUserSubmissionDir(courseID, assignmentID, user), shortSubmissionID)
}

func (this *backend) getSubmissionDirFromAssignment(assignment *model.Assignment, user string, shortSubmissionID string) string {
	return this.getSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), user, shortSubmissionID)
}

func (this *backend) getSubmissionDirFromResult(gradingInfo *model.GradingInfo) string {
	return this.getSubmissionDir(gradingInfo.CourseID, gradingInfo.AssignmentID, gradingInfo.User, gradingInfo.ShortID)
}

func (this *backend) getUserSubmissionDir(courseID string, assignmentID string, user string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), model.SUBMISSIONS_DIRNAME, assignmentID, user)
}

// Get the short ID of the most recent submission (or empty string if there are no submissions).
// The most recent submission has the largest ID for the user.
func (this *backend) getMostRecentSubmissionID(assignment *model.Assignment, email string) (string, error) {
	submissionsDir := this.getUserSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), email)
	if !util.PathExists(submissionsDir) {
		return "", nil
	}

	dirents, err := os.ReadDir(submissionsDir)
	if err != nil {
		return "", fmt.Errorf("Unable to read user submissions dir '%s': '%w'.", submissionsDir, err)
	}

	if len(dirents) == 0 {
		return "", nil
	}

	return dirents[len(dirents)-1].Name(), nil
}

func (this *backend) RemoveSubmission(assignment *model.Assignment, email string, shortSubmissionID string) (bool, error) {
	var err error

	if shortSubmissionID == "" {
		shortSubmissionID, err = this.getMostRecentSubmissionID(assignment, email)
		if err != nil {
			return false, fmt.Errorf("Failed to get most recent submission id: `%w`.", err)
		}
	}

	if shortSubmissionID == "" {
		return false, nil
	}

	submissionDir := this.getSubmissionDirFromAssignment(assignment, email, shortSubmissionID)

	if !util.PathExists(submissionDir) {
		return false, nil
	}

	err = util.RemoveDirent(submissionDir)
	if err != nil {
		wrappedErr := fmt.Errorf("Failed to remove submission '%s': '%w'", shortSubmissionID, err)
		return false, wrappedErr
	}

	return true, nil
}

func (this *backend) GetSubmissionAttempts(assignment *model.Assignment, email string) ([]*model.GradingResult, error) {
	submissions := make([]*model.GradingResult, 0)

	submissionsDir := this.getUserSubmissionDir(assignment.GetCourse().GetID(), assignment.GetID(), email)
	if !util.PathExists(submissionsDir) {
		return submissions, nil
	}

	dirents, err := os.ReadDir(submissionsDir)
	if err != nil {
		return nil, fmt.Errorf("Unable to read user submissions dir '%s': '%w'.", submissionsDir, err)
	}

	for _, dirent := range dirents {
		submission, err := this.GetSubmissionContents(assignment, email, dirent.Name())
		if err != nil {
			return nil, fmt.Errorf("Unable to get submission contents for %s: .", email)
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

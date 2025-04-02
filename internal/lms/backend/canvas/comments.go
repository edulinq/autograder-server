package canvas

import (
	"fmt"
	"time"

	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/util"
)

func (this *CanvasBackend) UpdateComments(assignmentID string, comments []*lmstypes.SubmissionComment) error {
	if assignmentID == "" {
		return fmt.Errorf("Cannot update comments, target assignment ID is empty.")
	}

	for i, comment := range comments {
		if i != 0 {
			time.Sleep(time.Duration(UPLOAD_SLEEP_TIME_SEC))
		}

		err := this.UpdateComment(assignmentID, comment)
		if err != nil {
			return fmt.Errorf("Failed on comment %d: '%w'.", i, err)
		}
	}

	return nil
}

func (this *CanvasBackend) UpdateComment(assignmentID string, comment *lmstypes.SubmissionComment) error {
	if assignmentID == "" {
		return fmt.Errorf("Cannot update comment, target assignment ID is empty.")
	}

	this.getAPILock()
	defer this.releaseAPILock()

	apiEndpoint := fmt.Sprintf(
		"/api/v1/courses/%s/assignments/%s/submissions/%s/comments/%s",
		this.CourseID, assignmentID, comment.Author, comment.ID)
	url := this.BaseURL + apiEndpoint

	form := make(map[string]string, 1)
	form["comment"] = comment.Text

	headers := this.standardHeaders()
	_, _, err := util.PutWithHeaders(url, form, headers)

	if err != nil {
		return fmt.Errorf("Failed to update comments: '%w'.", err)
	}

	return nil
}

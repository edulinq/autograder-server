package canvas

import (
    "fmt"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

func UpdateComments(canvasInfo *CanvasInstanceInfo, assignmentID string, comments []*CanvasSubmissionComment) error {
    for i, comment := range comments {
        if (i != 0) {
            time.Sleep(time.Duration(UPLOAD_SLEEP_TIME_SEC));
        }

        err := UpdateComment(canvasInfo, assignmentID, comment);
        if (err != nil) {
            return fmt.Errorf("Failed on comment %d: '%w'.", i, err);
        }
    }

    return nil;
}

func UpdateComment(canvasInfo *CanvasInstanceInfo, assignmentID string, comment *CanvasSubmissionComment) error {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions/%s/comments/%s",
        canvasInfo.CourseID, assignmentID, comment.Author, comment.ID);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    form := make(map[string]string, 1);
    form["comment"] = comment.Text;

    _, _, err := util.PutWithHeaders(url, form, headers);
    if (err != nil) {
        return fmt.Errorf("Failed to update comments: '%w'.", err);
    }

    return nil;
}

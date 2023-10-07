package canvas

import (
    "fmt"

    "github.com/eriq-augustine/autograder/util"
)

func FetchAssignment(canvasInfo *CanvasInstanceInfo, assignmentID string) (*CanvasAssignment, error) {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s",
        canvasInfo.CourseID, assignmentID);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    getAPILock(canvasInfo);
    body, _, err := util.GetWithHeaders(url, headers);
    releaseAPILock(canvasInfo);

    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch assignment: '%w'.", err);
    }

    var assignment CanvasAssignment;
    err = util.JSONFromString(body, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal assignment: '%w'.", err);
    }

    return &assignment, nil;
}

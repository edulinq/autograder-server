package canvas

import (
    "fmt"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/util"
)

func (this *CanvasBackend) FetchAssignment(assignmentID string) (*lmstypes.Assignment, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s",
        this.CourseID, assignmentID);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();
    body, _, err := common.GetWithHeaders(url, headers);

    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch assignment: '%w'.", err);
    }

    var assignment Assignment;
    err = util.JSONFromString(body, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal assignment: '%w'.", err);
    }

    return assignment.ToLMSType(), nil;
}

package canvas

import (
    "fmt"

    "github.com/eriq-augustine/autograder/util"
)

func FetchUsers(canvasInfo *CanvasInstanceInfo) ([]CanvasUserInfo, error) {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/users?per_page=%d",
        canvasInfo.CourseID, PAGE_SIZE);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    users := make([]CanvasUserInfo, 0);

    for (url != "") {
        body, responseHeaders, err := util.GetWithHeaders(url, headers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch users.");
        }

        var pageUsers []CanvasUserInfo;
        err = util.JSONFromString(body, &pageUsers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal users page: '%w'.", err);
        }

        users = append(users, pageUsers...);

        url = fetchNextCanvasLink(responseHeaders);
    }

    return users, nil;
}

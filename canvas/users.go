package canvas

import (
    "fmt"
    "net/url"

    "github.com/rs/zerolog/log"

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
        getAPILock(canvasInfo);
        body, responseHeaders, err := util.GetWithHeaders(url, headers);
        releaseAPILock(canvasInfo);

        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch users: '%w'.", err);
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

func FetchUser(canvasInfo *CanvasInstanceInfo, email string) (*CanvasUserInfo, error) {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/search_users?search_term=%s",
        canvasInfo.CourseID, url.QueryEscape(email));
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    getAPILock(canvasInfo);
    body, _, err := util.GetWithHeaders(url, headers);
    releaseAPILock(canvasInfo);

    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch user '%s': '%w'.", email, err);
    }

    var pageUsers []CanvasUserInfo;
    err = util.JSONFromString(body, &pageUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal user page: '%w'.", err);
    }

    if (len(pageUsers) != 1) {
        log.Warn().Str("email", email).Int("num-results", len(pageUsers)).Msg("Did not find exactly one matching user in canvas.");
        return nil, nil;
    }

    return &pageUsers[0], nil;
}

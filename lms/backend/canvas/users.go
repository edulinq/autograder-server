package canvas

import (
    "fmt"
    neturl "net/url"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
    "github.com/eriq-augustine/autograder/util"
)

func (this *CanvasBackend) FetchUsers() ([]*lmstypes.User, error) {
    return this.fetchUsers(false);
}

func (this *CanvasBackend) fetchUsers(rewriteLinks bool) ([]*lmstypes.User, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/users?include[]=enrollments&per_page=%d",
        this.CourseID, PAGE_SIZE);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();

    users := make([]*lmstypes.User, 0);

    for (url != "") {
        if (rewriteLinks) {
            parsed, err := neturl.Parse(url);
            if (err != nil) {
                return nil, fmt.Errorf("Failed to parse URL '%s': '%w'.", url, err);
            }

            url = fmt.Sprintf("%s%s?%s", this.BaseURL, parsed.Path, parsed.RawQuery);
        }

        body, responseHeaders, err := common.GetWithHeaders(url, headers);

        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch users: '%w'.", err);
        }

        var pageUsers []*User;
        err = util.JSONFromString(body, &pageUsers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal users page: '%w'.", err);
        }

        for _, user := range pageUsers {
            if (user == nil) {
                continue;
            }

            users = append(users, user.ToLMSType());
        }

        url = fetchNextCanvasLink(responseHeaders);
    }

    return users, nil;
}

func (this *CanvasBackend) FetchUser(email string) (*lmstypes.User, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/search_users?include[]=enrollments&search_term=%s",
        this.CourseID, neturl.QueryEscape(email));
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();
    body, _, err := common.GetWithHeaders(url, headers);

    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch user '%s': '%w'.", email, err);
    }

    var pageUsers []User;
    err = util.JSONFromString(body, &pageUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal user page: '%w'.", err);
    }

    if (len(pageUsers) != 1) {
        log.Warn().Str("email", email).Int("num-results", len(pageUsers)).Msg("Did not find exactly one matching user in canvas.");
        return nil, nil;
    }

    return pageUsers[0].ToLMSType(), nil;
}

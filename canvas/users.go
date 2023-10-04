package canvas

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const (
    PAGE_SIZE = 75
    HEADER_LINK = "Link";
)

func FetchUsers(canvasInfo *model.CanvasInfo) ([]model.CanvasUserInfo, error) {
    userListEndpoint := fmt.Sprintf("/api/v1/courses/%s/users", canvasInfo.CourseID);

    url := fmt.Sprintf("%s%s?per_page=%d", canvasInfo.BaseURL, userListEndpoint, PAGE_SIZE);
    headers := map[string][]string{
        "Authorization": []string{fmt.Sprintf("Bearer %s", canvasInfo.APIToken)},
        "Accept": []string{"application/json+canvas-string-ids"},
    };

    users := make([]model.CanvasUserInfo, 0);

    for (url != "") {
        body, responseHeaders, err := util.GetWithHeaders(url, headers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch users.");
        }

        var pageUsers []model.CanvasUserInfo;
        err = util.JSONFromString(body, &pageUsers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal users page: '%w'.", err);
        }

        users = append(users, pageUsers...);

        url = fetchNextCanvasLink(responseHeaders);
    }

    return users, nil;
}

// See if the response headers have a next link.
// Returns the link or an empty string.
func fetchNextCanvasLink(headers map[string][]string) string {
    values, ok := headers[HEADER_LINK];
    if (!ok) {
        return "";
    }

    for _, value := range values {
        links := strings.Split(value, ",");
        for _, link := range links {
            parts := strings.Split(link, ";");
            if (len(parts) < 2) {
                continue;
            }

            if (strings.TrimSpace(parts[1]) == `rel="next"`) {
                return strings.Trim(strings.TrimSpace(parts[0]), "<>");
            }
        }
    }

    return "";
}

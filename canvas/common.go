package canvas

import (
    "fmt"
    "strings"
    "sync"
)

const (
    PAGE_SIZE = 75
    HEADER_LINK = "Link";
)

// {string: *sync.Mutex}.
var apiLocks sync.Map;

func getAPILock(canvasInfo *CanvasInstanceInfo) {
    ensureAPILock(canvasInfo);
    lock, _ := apiLocks.Load(canvasInfo.APIToken);
    lock.(*sync.Mutex).Lock();
}

func releaseAPILock(canvasInfo *CanvasInstanceInfo) {
    ensureAPILock(canvasInfo);
    lock, _ := apiLocks.Load(canvasInfo.APIToken);
    lock.(*sync.Mutex).Unlock();
}

func ensureAPILock(canvasInfo *CanvasInstanceInfo) {
    apiLocks.LoadOrStore(canvasInfo.APIToken, &sync.Mutex{});
}

func standardHeaders(canvasInfo *CanvasInstanceInfo) map[string][]string {
    return map[string][]string{
        "Authorization": []string{fmt.Sprintf("Bearer %s", canvasInfo.APIToken)},
        "Accept": []string{"application/json+canvas-string-ids"},
    };
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

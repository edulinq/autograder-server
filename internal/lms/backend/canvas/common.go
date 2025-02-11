package canvas

import (
	"fmt"
	neturl "net/url"
	"strings"
	"time"

	"github.com/edulinq/autograder/internal/lockmanager"
)

const (
	PAGE_SIZE             int    = 75
	POST_PAGE_SIZE        int    = 75
	HEADER_LINK           string = "Link"
	UPLOAD_SLEEP_TIME_SEC        = int64(0.5 * float64(time.Second))
)

func (this *CanvasBackend) getAPILock() {
	lockmanager.Lock(this.getLockKey())
}

func (this *CanvasBackend) releaseAPILock() {
	lockmanager.Unlock(this.getLockKey())
}

// Lock based on the API token.
// Multiple courses can have the same token authenticating requests,
// and Canvas does rate limiting by token, not course.
func (this *CanvasBackend) getLockKey() string {
	return fmt.Sprintf("canvas::%s", this.APIToken)
}

func (this *CanvasBackend) standardHeaders() map[string][]string {
	return map[string][]string{
		"Authorization": []string{fmt.Sprintf("Bearer %s", this.APIToken)},
		"Accept":        []string{"application/json+canvas-string-ids"},
	}
}

// See if the response headers have a next link.
// Returns the link or an empty string.
func fetchNextCanvasLink(headers map[string][]string) string {
	values, ok := headers[HEADER_LINK]
	if !ok {
		return ""
	}

	for _, value := range values {
		links := strings.Split(value, ",")
		for _, link := range links {
			parts := strings.Split(link, ";")
			if len(parts) < 2 {
				continue
			}

			if strings.TrimSpace(parts[1]) == `rel="next"` {
				return strings.Trim(strings.TrimSpace(parts[0]), "<>")
			}
		}
	}

	return ""
}

// Rewrite a URL that appears in a LINK header for testing.
func (this *CanvasBackend) rewriteLink(url string) (string, error) {
	parsed, err := neturl.Parse(url)
	if err != nil {
		return "", fmt.Errorf("Failed to parse URL '%s': '%w'.", url, err)
	}

	url = fmt.Sprintf("%s%s?%s", this.BaseURL, parsed.Path, parsed.RawQuery)
	return url, nil
}

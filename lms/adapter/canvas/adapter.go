package canvas

import (
    "fmt"
    "strings"
)

type CanvasAdapter struct {
    CourseID string
    APIToken string
    BaseURL string
}

func NewAdapter(courseID string, apiToken string, baseURL string) (*CanvasAdapter, error) {
    if (courseID == "") {
        return nil, fmt.Errorf("Canvas course ID (course-id) cannot be empty.");
    }

    if (apiToken == "") {
        return nil, fmt.Errorf("Canvas API token (api-token) cannot be empty.");
    }

    if (baseURL == "") {
        return nil, fmt.Errorf("Canvas base URL (base-url) cannot be empty.");
    }

    baseURL = strings.TrimSuffix(baseURL, "/");

    adapter := CanvasAdapter{
        CourseID: courseID,
        APIToken: apiToken,
        BaseURL: baseURL,
    };

    return &adapter, nil;
}

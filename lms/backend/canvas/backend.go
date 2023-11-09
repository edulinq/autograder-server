package canvas

import (
    "fmt"
    "strings"
)

type CanvasBackend struct {
    CourseID string
    APIToken string
    BaseURL string
}

func NewBackend(canvasCourseID string, apiToken string, baseURL string) (*CanvasBackend, error) {
    if (canvasCourseID == "") {
        return nil, fmt.Errorf("Canvas course ID (course-id) cannot be empty.");
    }

    if (apiToken == "") {
        return nil, fmt.Errorf("Canvas API token (api-token) cannot be empty.");
    }

    if (baseURL == "") {
        return nil, fmt.Errorf("Canvas base URL (base-url) cannot be empty.");
    }

    baseURL = strings.TrimSuffix(baseURL, "/");

    backend := CanvasBackend{
        CourseID: canvasCourseID,
        APIToken: apiToken,
        BaseURL: baseURL,
    };

    return &backend, nil;
}

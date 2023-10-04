package model

import (
    "fmt"
    "strings"
)

type CanvasUserInfo struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Email string `json:"login_id"`
}

type CanvasGradeInfo struct {
    UserID string `json:"user_id"`
    Score string `json:"score"`
    Time string `json:"submitted_at"`
    Comments []CanvasSubmissionComment `json:"submission_comments"`
}

type CanvasSubmissionComment struct {
    ID string `json:"id"`
    Author string `json:"author_id"`
    Text string `json:"comment"`
    Time string `json:"edited_at"`
}

type CanvasInfo struct {
    CourseID string `json:"course-id"`
    APIToken string `json:"api-token"`
    BaseURL string `json:"base-url"`
}

func (this *CanvasInfo) Validate() error {
    if (this.CourseID == "") {
        return fmt.Errorf("Canvas course ID (course-id) cannot be empty.");
    }

    if (this.APIToken == "") {
        return fmt.Errorf("Canvas API token (api-token) cannot be empty.");
    }

    if (this.BaseURL == "") {
        return fmt.Errorf("Canvas base URL (base-url) cannot be empty.");
    }

    this.BaseURL = strings.TrimSuffix(this.BaseURL, "/");

    return nil;
}

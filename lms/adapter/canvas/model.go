package canvas

import (
    "time"

    "github.com/eriq-augustine/autograder/lms"
)

type User struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Email string `json:"login_id"`
}

type SubmissionScore struct {
    UserID string `json:"user_id"`
    Score float64 `json:"score"`
    Time time.Time `json:"submitted_at"`
    Comments []*SubmissionComment `json:"submission_comments"`
}

type SubmissionComment struct {
    ID string `json:"id"`
    Author string `json:"author_id"`
    Text string `json:"comment"`
    Time string `json:"edited_at"`
}

type Assignment struct {
    ID string `json:"id"`
    Name string `json:"name"`
    CourseID string `json:"course_id"`
    DueDate *time.Time `json:"due_at"`
    MaxPoints float64 `json:"points_possible"`
}

func (this *User) ToLMSType() *lms.User {
    return &lms.User{
        ID: this.ID,
        Name: this.Name,
        Email: this.Email,
    };
}

func (this *SubmissionScore) ToLMSType() *lms.SubmissionScore {
    comments := make([]*lms.SubmissionComment, 0, len(this.Comments));
    for _, comment := range this.Comments {
        comments = append(comments, comment.ToLMSType());
    }

    return &lms.SubmissionScore{
        UserID: this.UserID,
        Score: this.Score,
        Time: this.Time,
        Comments: comments,
    };
}

func (this *SubmissionComment) ToLMSType() *lms.SubmissionComment {
    return &lms.SubmissionComment{
        ID: this.ID,
        Author: this.Author,
        Text: this.Text,
        Time: this.Time,
    };
}

func (this *Assignment) ToLMSType() *lms.Assignment {
    return &lms.Assignment{
        ID: this.ID,
        Name: this.Name,
        CourseID: this.CourseID,
        DueDate: this.DueDate,
        MaxPoints: this.MaxPoints,
    };
}

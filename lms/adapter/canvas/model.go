package canvas

import (
    "time"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
)

type User struct {
    ID string `json:"id"`
    Name string `json:"name"`
    Email string `json:"login_id"`
    Enrollments []Enrollment `json:"enrollments"`
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

type Enrollment struct {
    ID string `json:"id"`
    Type string `json:"type"`
    EnrollmentState string `json:"enrollment_state"`
    Role string `json:"role"`
}

// Canvas enrollment to autograder role.
// Canvas has default enrollment "types" and then "roles" which may be the same
// as the type or custom.
var enrollmentToRoleMapping map[string]usr.UserRole = map[string]usr.UserRole{
    "ObserverEnrollment": usr.Other,
    "DesignerEnrollment": usr.Other,
    "StudentEnrollment": usr.Student,
    "TaEnrollment": usr.Grader,
    "TeacherEnrollment": usr.Owner,

    // Custom role.
    "TA - Site Manager": usr.Admin,
};

var roleToEnrollmentMapping map[usr.UserRole]string = map[usr.UserRole]string{
    usr.Other: "ObserverEnrollment",
    usr.Student: "StudentEnrollment",
    usr.Grader: "TaEnrollment",
    usr.Admin: "TA - Site Manager",
    usr.Owner: "TeacherEnrollment",
};

func (this *Enrollment) GetRole() usr.UserRole {
    typeRole := enrollmentToRoleMapping[this.Type];
    roleRole := enrollmentToRoleMapping[this.Role];

    return max(typeRole, roleRole);
}

func (this *User) GetRole() usr.UserRole {
    if (this.Enrollments == nil) {
        return usr.Other;
    }

    var maxRole usr.UserRole;
    for _, enrollment := range this.Enrollments {
        role := enrollment.GetRole();
        if (role > maxRole) {
            maxRole = role;
        }
    }

    return maxRole;
}

func (this *User) ToLMSType() *lms.User {
    return &lms.User{
        ID: this.ID,
        Name: this.Name,
        Email: this.Email,
        Role: this.GetRole(),
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

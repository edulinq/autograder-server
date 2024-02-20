package canvas

import (
    "time"

    "github.com/edulinq/autograder/lms/lmstypes"
    "github.com/edulinq/autograder/model"
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
    CanvasCourseID string `json:"course_id"`
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
var enrollmentToRoleMapping map[string]model.UserRole = map[string]model.UserRole{
    "ObserverEnrollment": model.RoleOther,
    "DesignerEnrollment": model.RoleOther,
    "StudentEnrollment": model.RoleStudent,
    "TaEnrollment": model.RoleGrader,
    "TeacherEnrollment": model.RoleOwner,

    // Custom role.
    "TA - Site Manager": model.RoleAdmin,
};

var roleToEnrollmentMapping map[model.UserRole]string = map[model.UserRole]string{
    model.RoleOther: "ObserverEnrollment",
    model.RoleStudent: "StudentEnrollment",
    model.RoleGrader: "TaEnrollment",
    model.RoleAdmin: "TA - Site Manager",
    model.RoleOwner: "TeacherEnrollment",
};

func (this *Enrollment) GetRole() model.UserRole {
    typeRole := enrollmentToRoleMapping[this.Type];
    roleRole := enrollmentToRoleMapping[this.Role];

    return max(typeRole, roleRole);
}

func (this *User) GetRole() model.UserRole {
    if (this.Enrollments == nil) {
        return model.RoleOther;
    }

    var maxRole model.UserRole;
    for _, enrollment := range this.Enrollments {
        role := enrollment.GetRole();
        if (role > maxRole) {
            maxRole = role;
        }
    }

    return maxRole;
}

func (this *User) ToLMSType() *lmstypes.User {
    return &lmstypes.User{
        ID: this.ID,
        Name: this.Name,
        Email: this.Email,
        Role: this.GetRole(),
    };
}

func (this *SubmissionScore) ToLMSType() *lmstypes.SubmissionScore {
    comments := make([]*lmstypes.SubmissionComment, 0, len(this.Comments));
    for _, comment := range this.Comments {
        comments = append(comments, comment.ToLMSType());
    }

    return &lmstypes.SubmissionScore{
        UserID: this.UserID,
        Score: this.Score,
        Time: this.Time,
        Comments: comments,
    };
}

func (this *SubmissionComment) ToLMSType() *lmstypes.SubmissionComment {
    return &lmstypes.SubmissionComment{
        ID: this.ID,
        Author: this.Author,
        Text: this.Text,
        Time: this.Time,
    };
}

func (this *Assignment) ToLMSType() *lmstypes.Assignment {
    return &lmstypes.Assignment{
        ID: this.ID,
        Name: this.Name,
        LMSCourseID: this.CanvasCourseID,
        DueDate: this.DueDate,
        MaxPoints: this.MaxPoints,
    };
}

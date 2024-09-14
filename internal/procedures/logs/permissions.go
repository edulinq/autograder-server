package logs

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func checkPermissions(query log.ParsedLogQuery, user *model.ServerUser) error {
	if user == nil {
		return fmt.Errorf("Cannot check log permissions with a nil user.")
	}

	// Server admins can do what they want.
	if user.Role >= model.ServerRoleAdmin {
		return nil
	}

	// A user may query for their own records.
	if (query.UserEmail != "") && (query.UserEmail == user.Email) {
		return nil
	}

	// A course grader may query course records.
	queryCourseID, _ := common.ValidateID(query.CourseID)
	if (queryCourseID != "") && (user.GetCourseRole(queryCourseID) >= model.CourseRoleGrader) {
		return nil
	}

	return fmt.Errorf("User does not meet conditions for query.")
}

package status

import (
	"fmt"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type SetRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	// Indicates whether the course should be active or inactive.
	Active bool `json:"active"`

	// Optional message to include with the status change.
	Message string `json:"message"`

	// If the status already exists for the user, allows an overwrite.
	Force bool `json:"force"`
}

type SetResponse struct {
	Status *model.CourseStatus `json:"status"`
}

// Set the status of a course, where each user can have at most one status.
func HandleSet(request *SetRequest) (*SetResponse, *core.APIError) {
	owner := request.ServerUser.Email

	existingStatus, ok := request.Course.Statuses[owner]
	if ok && !request.Force {
		return nil, core.NewBadRequestError("-646", request,
			fmt.Sprintf("Course active/inactive status already exists for this user (set at %s), use the force option to overwrite.",
				existingStatus.SetTime.SafeString()))
	}

	status := &model.CourseStatus{
		Active:  request.Active,
		Source:  determineSource(request.ServerUser.Role),
		Owner:   owner,
		Message: request.Message,
		SetTime: request.Timestamp,
	}

	request.Course.Statuses[owner] = status

	err := db.SaveCourse(request.Course)
	if err != nil {
		return nil, core.NewInternalError("-647", request, "Failed to save course status.").Err(err)
	}

	log.Info(
		"Course status set.",
		request.Course,
		request.ServerUser,
		log.NewAttr("message", request.Message),
		log.NewAttr("active", request.Active),
		log.NewAttr("source", status.Source),
	)

	return &SetResponse{Status: status}, nil
}

// Finds the highest source using the user's server role.
// The request permissions require the caller to be at least a course admin.
func determineSource(serverRole model.ServerUserRole) model.StatusSource {
	if serverRole >= model.ServerRoleRoot {
		return model.StatusSourceRoot
	}

	if serverRole >= model.ServerRoleAdmin {
		return model.StatusSourceServer
	}

	return model.StatusSourceCourse
}

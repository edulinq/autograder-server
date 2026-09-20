package status

import (
	"fmt"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type RemoveRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	// Email of status owner to remove. Defaults to the caller.
	TargetOwner string `json:"target-owner"`

	// If true, ignores the target owner and removes all statuses the caller has permission to remove.
	Clear bool `json:"clear"`

	// Optional log message to include with status removal.
	Message string `json:"message"`
}

type RemoveResponse struct {
	Removed map[string]model.CourseStatus `json:"removed"`
}

// Remove one or more course statuses based on the caller's permissions.
func HandleRemove(request *RemoveRequest) (*RemoveResponse, *core.APIError) {
	callerSource := determineSource(request.ServerUser.Role)
	courseStatuses := request.Course.Statuses

	removed := make(map[string]model.CourseStatus)

	if request.Clear {
		for owner, status := range courseStatuses {
			if callerSource >= status.Source {
				delete(courseStatuses, owner)
				removed[owner] = *status
			}
		}
	} else {
		target := request.TargetOwner
		if target == "" {
			target = request.ServerUser.Email
		}

		targetStatus, ok := courseStatuses[target]
		if ok {
			if callerSource < targetStatus.Source {
				return nil, core.NewBadRequestError("-648", request,
					fmt.Sprintf("Cannot remove this course's active/inactive status because it was set by a higher privileged source (%s) than yours (%s).",
						targetStatus.Source, callerSource))
			}

			delete(courseStatuses, target)
			removed[target] = *targetStatus
		}
	}

	err := db.SaveCourse(request.Course)
	if err != nil {
		return nil, core.NewInternalError("-649", request, "Failed to save course status.").Err(err)
	}

	log.Info(
		"Course status removal.",
		request.Course,
		request.ServerUser,
		log.NewAttr("message", request.Message),
		log.NewAttr("removed", removed),
	)

	return &RemoveResponse{Removed: removed}, nil
}

package lms

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/lms"
)

type UserGetRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	TargetUser core.TargetCourseUser `json:"target-email"`
}

type UserGetResponse struct {
	FoundAGUser  bool                 `json:"found-autograder-user"`
	FoundLMSUser bool                 `json:"found-lms-user"`
	User         *core.CourseUserInfo `json:"user"`
}

// Get information for an LMS user.
func HandleUserGet(request *UserGetRequest) (*UserGetResponse, *core.APIError) {
	if request.Course.GetLMSAdapter() == nil {
		return nil, core.NewBadRequestError("-401", &request.APIRequest, "Course is not linked to an LMS.").
			Course(request.Course.GetID())
	}

	response := UserGetResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundAGUser = true
	response.User = core.NewCourseUserInfo(request.TargetUser.User)

	lmsUser, err := lms.FetchUser(request.Course, string(request.TargetUser.Email))
	if err != nil {
		return nil, core.NewInternalError("-402", &request.APIRequestCourseUserContext,
			"Failed to fetch LMS user.").Err(err).Add("target-user", string(request.TargetUser.Email))
	}

	if lmsUser == nil {
		return &response, nil
	}

	response.FoundLMSUser = true
	response.User.Name = lmsUser.Name
	response.User.LMSID = lmsUser.ID

	return &response, nil
}

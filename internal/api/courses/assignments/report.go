package assignments

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/report"
)

type CourseReportRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
}

type CourseReportResponse struct {
	CourseReport *report.CourseScoringReport `json:"course-report"`
}

// Fetch course report.
func HandleCourseReport(request *CourseReportRequest) (*CourseReportResponse, *core.APIError) {
	courseReport, err := report.GetCourseScoringReport(request.Course)
	if err != nil {
		return nil, core.NewInternalError("-700", request, "Unable to fetch course report.").Err(err)
	}

	response := CourseReportResponse{
		CourseReport: courseReport,
	}

	return &response, nil
}

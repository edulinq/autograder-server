package submission

import (
	// "fmt"

	"github.com/edulinq/autograder/api/core"
	//"github.com/edulinq/autograder/model"
	"github.com/edulinq/autograder/report"
)

// permission erros
// non-existing course

type FetchCourseReportRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin
	
}

type FetchCourseReportResponse struct {
	CourseReport *report.CourseScoringReport `json:"course-report"`
}

func HandleFetchCourseReport(request *FetchCourseReportRequest) (*FetchCourseReportResponse, *core.APIError) {

	response := FetchCourseReportResponse{};

	gettingCourseReport, err := report.GetCourseScoringReport(request.Course);


	if err != nil {
        return nil, core.NewInternalError("-608", &request.APIRequestCourseUserContext, "Failed to get course report.").
            Err(err).Course(request.CourseID);
    }

	response.CourseReport = gettingCourseReport;

	return &response, nil
}
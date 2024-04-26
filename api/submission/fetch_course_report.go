package submission

import (
    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/report"
)

type FetchCourseReportRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
}

type FetchCourseReportResponse struct {
    CourseReport *report.CourseScoringReport `json:"course-report"`
}

func HandleFetchCourseReport(request *FetchCourseReportRequest) (*FetchCourseReportResponse, *core.APIError) {
    courseReport, err := report.GetCourseScoringReport(request.Course);
    if err != nil {
        return nil, core.NewInternalError("-608", &request.APIRequestCourseUserContext, "Failed to get course report.").
            Err(err).Course(request.CourseID);
    }

    return &FetchCourseReportResponse{courseReport}, nil
}

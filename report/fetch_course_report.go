package report

import (
    "github.com/edulinq/autograder/api/core"
)

type FetchCourseReportRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
}

type FetchCourseReportResponse struct {
    CourseReport *CourseScoringReport `json:"course-report"`
}

func HandleFetchCourseReport(request *FetchCourseReportRequest) (*FetchCourseReportResponse, *core.APIError) {
    courseReport, err := GetCourseScoringReport(request.Course);
    if err != nil {
        return nil, core.NewInternalError("-608", &request.APIRequestCourseUserContext, "Failed to get course report.").
            Err(err).Course(request.CourseID);
    }

    return &FetchCourseReportResponse{courseReport}, nil
}

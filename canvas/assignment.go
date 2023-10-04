package canvas

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

func FetchAssignmentGrades(canvasInfo *model.CanvasInfo, assignmentID string) ([]model.CanvasGradeInfo, error) {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions?per_page=%d&include[]=submission_comments",
        canvasInfo.CourseID, assignmentID, PAGE_SIZE);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    grades := make([]model.CanvasGradeInfo, 0);

    for (url != "") {
        body, responseHeaders, err := util.GetWithHeaders(url, headers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch grades.");
        }

        var pageGrades []model.CanvasGradeInfo;
        err = util.JSONFromString(body, &pageGrades);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal grades page: '%w'.", err);
        }

        grades = append(grades, pageGrades...);

        url = fetchNextCanvasLink(responseHeaders);
    }

    return grades, nil;
}

package canvas

import (
    "fmt"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

func FetchAssignmentGrades(canvasInfo *CanvasInstanceInfo, assignmentID string) ([]*CanvasGradeInfo, error) {
    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions?per_page=%d&include[]=submission_comments",
        canvasInfo.CourseID, assignmentID, PAGE_SIZE);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    grades := make([]*CanvasGradeInfo, 0);

    for (url != "") {
        getAPILock(canvasInfo);
        body, responseHeaders, err := util.GetWithHeaders(url, headers);
        releaseAPILock(canvasInfo);

        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch grades.");
        }

        var pageGrades []*CanvasGradeInfo;
        err = util.JSONFromString(body, &pageGrades);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal grades page: '%w'.", err);
        }

        grades = append(grades, pageGrades...);

        url = fetchNextCanvasLink(responseHeaders);
    }

    return grades, nil;
}

func UpdateAssignmentGrades(canvasInfo *CanvasInstanceInfo, assignmentID string, grades []*CanvasGradeInfo) error {
    for page := 0; (page * POST_PAGE_SIZE) < len(grades); page++ {
        startIndex := page * POST_PAGE_SIZE;
        endIndex := min(len(grades), ((page + 1) * POST_PAGE_SIZE));

        if (page != 0) {
            time.Sleep(time.Duration(UPLOAD_SLEEP_TIME_SEC));
        }

        err := updateAssignmentGrades(canvasInfo, assignmentID, grades[startIndex:endIndex]);
        if (err != nil) {
            return fmt.Errorf("Failed on page %d: '%w'.", page, err);
        }
    }

    return nil;
}

func updateAssignmentGrades(canvasInfo *CanvasInstanceInfo, assignmentID string, grades []*CanvasGradeInfo) error {
    if (len(grades) > POST_PAGE_SIZE) {
        return fmt.Errorf("Too many grade upload requests at once. Found %d, max %d.", len(grades), POST_PAGE_SIZE);
    }

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions/update_grades",
        canvasInfo.CourseID, assignmentID);
    url := canvasInfo.BaseURL + apiEndpoint;

    headers := standardHeaders(canvasInfo);

    form := make(map[string]string);

    for _, gradeInfo := range grades {
        form[fmt.Sprintf("grade_data[%s][posted_grade]", gradeInfo.UserID)] = util.FloatToStr(gradeInfo.Score);

        if (len(gradeInfo.Comments) > 1) {
            return fmt.Errorf("Grades to upload can have at most one comment. Student '%s' for assignment '%s' has %d.", gradeInfo.UserID, assignmentID, len(gradeInfo.Comments));
        }

        for _, comment := range gradeInfo.Comments {
            form[fmt.Sprintf("grade_data[%s][text_comment]", gradeInfo.UserID)] = comment.Text;
        }
    }

    getAPILock(canvasInfo);
    _, _, err := util.PostWithHeaders(url, form, headers);
    releaseAPILock(canvasInfo);

    if (err != nil) {
        return fmt.Errorf("Failed to upload grades: '%w'.", err);
    }

    return nil;
}

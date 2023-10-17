package canvas

import (
    "fmt"
    "time"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/util"
)

func (this *CanvasAdapter) FetchAssignmentScores(assignmentID string) ([]*lms.SubmissionScore, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions?per_page=%d&include[]=submission_comments",
        this.CourseID, assignmentID, PAGE_SIZE);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();

    scores := make([]*lms.SubmissionScore, 0);

    for (url != "") {
        body, responseHeaders, err := util.GetWithHeaders(url, headers);

        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch scores.");
        }

        var pageScores []*SubmissionScore;
        err = util.JSONFromString(body, &pageScores);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal scores page: '%w'.", err);
        }

        for _, score := range pageScores {
            if (score == nil) {
                continue;
            }

            scores = append(scores, score.ToLMSType());
        }

        url = fetchNextCanvasLink(responseHeaders);
    }

    return scores, nil;
}

func (this *CanvasAdapter) UpdateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error {
    for page := 0; (page * POST_PAGE_SIZE) < len(scores); page++ {
        startIndex := page * POST_PAGE_SIZE;
        endIndex := min(len(scores), ((page + 1) * POST_PAGE_SIZE));

        if (page != 0) {
            time.Sleep(time.Duration(UPLOAD_SLEEP_TIME_SEC));
        }

        err := this.updateAssignmentScores(assignmentID, scores[startIndex:endIndex]);
        if (err != nil) {
            return fmt.Errorf("Failed on page %d: '%w'.", page, err);
        }
    }

    return nil;
}

func (this *CanvasAdapter) updateAssignmentScores(assignmentID string, scores []*lms.SubmissionScore) error {
    this.getAPILock();
    defer this.releaseAPILock();

    if (len(scores) > POST_PAGE_SIZE) {
        return fmt.Errorf("Too many score upload requests at once. Found %d, max %d.", len(scores), POST_PAGE_SIZE);
    }

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions/update_grades",
        this.CourseID, assignmentID);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();

    form := make(map[string]string);

    for _, score := range scores {
        form[fmt.Sprintf("grade_data[%s][posted_grade]", score.UserID)] = util.FloatToStr(score.Score);

        if (len(score.Comments) > 1) {
            return fmt.Errorf("Scores to upload can have at most one comment. Student '%s' for assignment '%s' has %d.", score.UserID, assignmentID, len(score.Comments));
        }

        for _, comment := range score.Comments {
            form[fmt.Sprintf("grade_data[%s][text_comment]", score.UserID)] = comment.Text;
        }
    }

    _, _, err := util.PostWithHeaders(url, form, headers);
    if (err != nil) {
        return fmt.Errorf("Failed to upload scores: '%w'.", err);
    }

    return nil;
}

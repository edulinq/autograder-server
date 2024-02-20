package canvas

import (
    "fmt"
    "time"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/lms/lmstypes"
    "github.com/edulinq/autograder/util"
)

func (this *CanvasBackend) FetchAssignmentScore(assignmentID string, userID string) (*lmstypes.SubmissionScore, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions/%s?include[]=submission_comments",
        this.CourseID, assignmentID, userID);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();
    body, _, err := common.GetWithHeaders(url, headers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch score.");
    }

    var rawScore SubmissionScore;
    err = util.JSONFromString(body, &rawScore);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal raw scores: '%w'.", err);
    }

    return rawScore.ToLMSType(), nil;
}

func (this *CanvasBackend) FetchAssignmentScores(assignmentID string) ([]*lmstypes.SubmissionScore, error) {
    return this.fetchAssignmentScores(assignmentID, false);
}

func (this *CanvasBackend) fetchAssignmentScores(assignmentID string, rewriteLinks bool) ([]*lmstypes.SubmissionScore, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/assignments/%s/submissions?per_page=%d&include[]=submission_comments",
        this.CourseID, assignmentID, PAGE_SIZE);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();

    scores := make([]*lmstypes.SubmissionScore, 0);

    for (url != "") {
        var err error;

        if (rewriteLinks) {
            url, err = this.rewriteLink(url);
            if (err != nil) {
                return nil, err;
            }
        }

        body, responseHeaders, err := common.GetWithHeaders(url, headers);

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

func (this *CanvasBackend) UpdateAssignmentScores(assignmentID string, scores []*lmstypes.SubmissionScore) error {
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

func (this *CanvasBackend) updateAssignmentScores(assignmentID string, scores []*lmstypes.SubmissionScore) error {
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

    _, _, err := common.PostWithHeaders(url, form, headers);
    if (err != nil) {
        return fmt.Errorf("Failed to upload scores: '%w'.", err);
    }

    return nil;
}

package model

import (
    "fmt"
    "strings"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func (this *Assignment) FullScoringAndUpload(dryRun bool) error {
    if (this.Course.CanvasInstanceInfo == nil) {
        return fmt.Errorf("Assignment's course has no Canvas info associated with it.");
    }

    users, err := this.Course.GetUsers();
    if (err != nil) {
        return fmt.Errorf("Failed to fetch autograder users: '%w'.", err);
    }

    canvasGrades, err := canvas.FetchAssignmentGrades(this.Course.CanvasInstanceInfo, this.CanvasID);
    if (err != nil) {
        return fmt.Errorf("Could not fetch Canvas grades: '%w'.", err);
    }

    scoringInfos, err := this.GetScoringInfo(users, true);
    if (err != nil) {
        return fmt.Errorf("Failed to get scoring information: '%w'.", err);
    }

    err = this.LatePolicy.Apply(this, users, scoringInfos, dryRun);
    if (err != nil) {
        return fmt.Errorf("Failed to apply late policy: '%w'.", err);
    }

    err = computeFinalScores(this, users, scoringInfos, canvasGrades, dryRun);
    if (err != nil) {
        return fmt.Errorf("Failed to apply late policy: '%w'.", err);
    }

    return nil;
}

// Get all the recent submission summaries (via GetAllRecentSubmissionSummaries()),
// and convert them to scoring info structs so they can be properly scored/uploaded.
func (this *Assignment) GetScoringInfo(users map[string]*usr.User, onlyStudents bool) (map[string]*artifact.ScoringInfo, error) {
    paths, err := this.GetAllRecentSubmissionSummaries(users);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to load submission summaries: '%w'.", err);
    }

    results := make(map[string]*artifact.ScoringInfo, len(paths));

    for username, path := range paths {
        if (path == "") {
            continue;
        }

        if (onlyStudents && (users[username].Role != usr.Student)) {
            continue;
        }

        var summary artifact.SubmissionSummary;
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, fmt.Errorf("Unable to load submission summary from path '%s': '%w'.", path, err);
        }

        results[username] = summary.GetScoringInfo();
    }

    return results, nil;
}

func computeFinalScores(
        assignment *Assignment, users map[string]*usr.User,
        scoringInfos map[string]*artifact.ScoringInfo, canvasGrades []*canvas.CanvasGradeInfo,
        dryRun bool) error {
    var err error;

    // First, look through canvas comments for locks and autograder notes.
    locks, existingComments, err := parseComments(canvasGrades);
    if (err != nil) {
        return err;
    }

    // Next, create the grades that will actually be uploaded and the comments that will be updated..
    finalGrades, commentsToUpdate := filterFinalScores(users, scoringInfos, locks, existingComments);

    // Upload the grades.
    if (dryRun) {
        log.Info().Str("assignment", assignment.ID).Any("grades", finalGrades).Msg("Dry Run: Skipping upload of final grades.");
    } else {
        err = canvas.UpdateAssignmentGrades(assignment.Course.CanvasInstanceInfo, assignment.CanvasID, finalGrades);
        if (err != nil) {
            return fmt.Errorf("Failed to upload final scores: '%w'.", err);
        }
    }

    // Update the comments.
    if (dryRun) {
        log.Info().Str("assignment", assignment.ID).Any("comments", commentsToUpdate).Msg("Dry Run: Skipping update of final comments.");
    } else {
        err = canvas.UpdateComments(assignment.Course.CanvasInstanceInfo, assignment.CanvasID, commentsToUpdate);
        if (err != nil) {
            return fmt.Errorf("Failed to update final comments: '%w'.", err);
        }
    }

    return nil;
}

func parseComments(canvasGrades []*canvas.CanvasGradeInfo) (map[string]bool, map[string]*artifact.ScoringInfo, error) {
    locks := make(map[string]bool);
    existingComments := make(map[string]*artifact.ScoringInfo);

    for _, canvasGradeInfo := range canvasGrades {
        for _, comment := range canvasGradeInfo.Comments {
            text := strings.ToLower(comment.Text);
            if (strings.Contains(text, canvas.LOCK_COMMENT)) {
                locks[canvasGradeInfo.UserID] = true;
            } else if (strings.Contains(text, common.AUTOGRADER_COMMENT_IDENTITY_KEY)) {
                var scoringInfo artifact.ScoringInfo;
                err := util.JSONFromString(comment.Text, &scoringInfo);
                if (err != nil) {
                    return nil, nil, fmt.Errorf("Could not unmarshall Canvas comment %s (%s) into a scoring info: '%w'.", comment.ID, comment.Text, err);
                }
                scoringInfo.CanvasCommentID = comment.ID;
                scoringInfo.CanvasCommentAuthorID = comment.Author;

                existingComments[canvasGradeInfo.UserID] = &scoringInfo;

                // Scoring infos can also lock grades.
                if (scoringInfo.Lock) {
                    locks[canvasGradeInfo.UserID] = true;
                }
            }
        }
    }

    return locks, existingComments, nil;
}

func filterFinalScores(
        users map[string]*usr.User, scoringInfos map[string]*artifact.ScoringInfo,
        locks map[string]bool, existingComments map[string]*artifact.ScoringInfo,
        ) ([]*canvas.CanvasGradeInfo, []*canvas.CanvasSubmissionComment) {
    finalGrades := make([]*canvas.CanvasGradeInfo, 0);
    commentsToUpdate := make([]*canvas.CanvasSubmissionComment, 0);

    for email, scoringInfo := range scoringInfos {
        user := users[email];
        if (user == nil) {
            log.Warn().Str("user", email).Msg("User does not exist, skipping grade upload.");
            continue;
        }

        // This scoring is invalid, skip it.
        if (scoringInfo.Reject) {
            continue;
        }

        // Skip users that do not have a canvas id.
        if (user.CanvasID == "") {
            log.Warn().Str("user", email).Msg("User does not have a Canvas ID, skipping grade upload.");
            continue;
        }

        // This score is locked, skip it.
        if (locks[user.CanvasID]) {
            continue;
        }

        scoringInfo.UploadTime = time.Now();

        // Check the existing comment last so we can decide if this comment needs to be updated.
        existingComment := existingComments[user.CanvasID];
        if (existingComment != nil) {
            // If this user has an existing comment, then we may skip this upload if submission IDs match.
            if (existingComment.ID == scoringInfo.ID) {
                log.Trace().Str("user", email).Str("submittion-id", existingComment.ID).Msg("User's submission/grade is up-to-date.");
                continue;
            }
        }

        // This scoring is valid and different than the last one.

        // Existing comments are updated, new comments are posted with the grade.
        var uploadComments []canvas.CanvasSubmissionComment = nil;

        if (existingComment != nil) {
            scoringInfo.CanvasCommentID = existingComment.CanvasCommentID;
            scoringInfo.CanvasCommentAuthorID = existingComment.CanvasCommentAuthorID;

            commentsToUpdate = append(commentsToUpdate, &canvas.CanvasSubmissionComment{
                ID: scoringInfo.CanvasCommentID,
                Author: scoringInfo.CanvasCommentAuthorID,
                Text: util.MustToJSON(scoringInfo),
            });
        } else {
            uploadComments = []canvas.CanvasSubmissionComment{
                canvas.CanvasSubmissionComment{
                    Text: util.MustToJSON(scoringInfo),
                },
            };
        }

        canvasGrade := canvas.CanvasGradeInfo{
            UserID: user.CanvasID,
            Score: scoringInfo.Score,
            Time: scoringInfo.SubmissionTime,
            Comments: uploadComments,
        }

        finalGrades = append(finalGrades, &canvasGrade);
    }

    return finalGrades, commentsToUpdate;
}

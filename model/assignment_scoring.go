package model

import (
    "fmt"
    "strings"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const LOCK_COMMENT string = "__lock__";

func (this *Assignment) FullScoringAndUpload(dryRun bool) error {
    if (this.Course.GetLMSAdapter() == nil) {
        return fmt.Errorf("Assignment's course has no LMS info associated with it.");
    }

    users, err := this.Course.GetUsers();
    if (err != nil) {
        return fmt.Errorf("Failed to fetch autograder users: '%w'.", err);
    }

    lmsScores, err := this.Course.GetLMSAdapter().FetchAssignmentScores(this.LMSID);
    if (err != nil) {
        return fmt.Errorf("Could not fetch LMS grades: '%w'.", err);
    }

    scoringInfos, err := this.GetScoringInfo(users, true);
    if (err != nil) {
        return fmt.Errorf("Failed to get scoring information: '%w'.", err);
    }

    err = this.LatePolicy.Apply(this, users, scoringInfos, dryRun);
    if (err != nil) {
        return fmt.Errorf("Failed to apply late policy: '%w'.", err);
    }

    err = computeFinalScores(this, users, scoringInfos, lmsScores, dryRun);
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
        scoringInfos map[string]*artifact.ScoringInfo, lmsScores []*lms.SubmissionScore,
        dryRun bool) error {
    var err error;

    // First, look through comments for locks and autograder notes.
    locks, existingComments, err := parseComments(lmsScores);
    if (err != nil) {
        return err;
    }

    // Next, create the grades that will actually be uploaded and the comments that will be updated..
    finalScores, commentsToUpdate := filterFinalScores(users, scoringInfos, locks, existingComments);

    // Upload the grades.
    if (dryRun) {
        log.Info().Str("assignment", assignment.GetID()).Any("grades", finalScores).Msg("Dry Run: Skipping upload of final grades.");
    } else {
        err = assignment.GetCourse().GetLMSAdapter().UpdateAssignmentScores(assignment.GetLMSID(), finalScores);
        if (err != nil) {
            return fmt.Errorf("Failed to upload final scores: '%w'.", err);
        }
    }

    // Update the comments.
    if (dryRun) {
        log.Info().Str("assignment", assignment.GetID()).Any("comments", commentsToUpdate).Msg("Dry Run: Skipping update of final comments.");
    } else {
        err = assignment.GetCourse().GetLMSAdapter().UpdateComments(assignment.GetLMSID(), commentsToUpdate);
        if (err != nil) {
            return fmt.Errorf("Failed to update final comments: '%w'.", err);
        }
    }

    return nil;
}

func parseComments(lmsScores []*lms.SubmissionScore) (map[string]bool, map[string]*artifact.ScoringInfo, error) {
    locks := make(map[string]bool);
    existingComments := make(map[string]*artifact.ScoringInfo);

    for _, lmsScore := range lmsScores {
        for _, comment := range lmsScore.Comments {
            text := strings.ToLower(comment.Text);
            if (strings.Contains(text, LOCK_COMMENT)) {
                locks[lmsScore.UserID] = true;
            } else if (strings.Contains(text, common.AUTOGRADER_COMMENT_IDENTITY_KEY)) {
                var scoringInfo artifact.ScoringInfo;
                err := util.JSONFromString(comment.Text, &scoringInfo);
                if (err != nil) {
                    return nil, nil, fmt.Errorf("Could not unmarshall LMS comment %s (%s) into a scoring info: '%w'.", comment.ID, comment.Text, err);
                }
                scoringInfo.LMSCommentID = comment.ID;
                scoringInfo.LMSCommentAuthorID = comment.Author;

                existingComments[lmsScore.UserID] = &scoringInfo;

                // Scoring infos can also lock grades.
                if (scoringInfo.Lock) {
                    locks[lmsScore.UserID] = true;
                }
            }
        }
    }

    return locks, existingComments, nil;
}

func filterFinalScores(
        users map[string]*usr.User, scoringInfos map[string]*artifact.ScoringInfo,
        locks map[string]bool, existingComments map[string]*artifact.ScoringInfo,
        ) ([]*lms.SubmissionScore, []*lms.SubmissionComment) {
    finalScores := make([]*lms.SubmissionScore, 0);
    commentsToUpdate := make([]*lms.SubmissionComment, 0);

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

        // Skip users that do not have an LMS id.
        if (user.LMSID == "") {
            log.Warn().Str("user", email).Msg("User does not have an LMS ID, skipping grade upload.");
            continue;
        }

        // This score is locked, skip it.
        if (locks[user.LMSID]) {
            continue;
        }

        scoringInfo.UploadTime = time.Now();

        // Check the existing comment last so we can decide if this comment needs to be updated.
        existingComment := existingComments[user.LMSID];
        if (existingComment != nil) {
            // If this user has an existing comment, then we may skip this upload if submission IDs match.
            if (existingComment.ID == scoringInfo.ID) {
                log.Trace().Str("user", email).Str("submittion-id", existingComment.ID).Msg("User's submission/grade is up-to-date.");
                continue;
            }
        }

        // This scoring is valid and different than the last one.

        // Existing comments are updated, new comments are posted with the grade.
        var uploadComments []*lms.SubmissionComment = nil;

        if (existingComment != nil) {
            scoringInfo.LMSCommentID = existingComment.LMSCommentID;
            scoringInfo.LMSCommentAuthorID = existingComment.LMSCommentAuthorID;

            commentsToUpdate = append(commentsToUpdate, &lms.SubmissionComment{
                ID: scoringInfo.LMSCommentID,
                Author: scoringInfo.LMSCommentAuthorID,
                Text: util.MustToJSON(scoringInfo),
            });
        } else {
            uploadComments = []*lms.SubmissionComment{
                &lms.SubmissionComment{
                    Text: util.MustToJSON(scoringInfo),
                },
            };
        }

        lmsScore := lms.SubmissionScore{
            UserID: user.LMSID,
            Score: scoringInfo.Score,
            Time: scoringInfo.SubmissionTime,
            Comments: uploadComments,
        }

        finalScores = append(finalScores, &lmsScore);
    }

    return finalScores, commentsToUpdate;
}

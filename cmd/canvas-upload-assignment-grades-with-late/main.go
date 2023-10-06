package main

import (
    "fmt"
    "strings"

    "github.com/alecthomas/kong"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var args struct {
    config.ConfigArgs
    AssignmentPath string `help:"Path to assignment JSON file." arg:"" type:"existingfile"`
    DryRun bool `help:"Do not actually upload the grades, just state what you would do." default:"false"`
}

func main() {
    kong.Parse(&args,
        kong.Description("Upload grades after late polices have been applied for an assignment to canvas from local submissions."),
    );

    err := config.HandleConfigArgs(args.ConfigArgs);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load config options.");
    }

    assignment := model.MustLoadAssignmentConfig(args.AssignmentPath);
    if (assignment.CanvasID == "") {
        log.Fatal().Msg("Assignment has no Canvas ID.");
    }

    if (assignment.Course.CanvasInstanceInfo == nil) {
        log.Fatal().Msg("Assignment's course has no Canvas info associated with it.");
    }

    users, err := assignment.Course.GetUsers();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to fetch autograder users.");
    }

    canvasGrades, err := canvas.FetchAssignmentGrades(assignment.Course.CanvasInstanceInfo, assignment.CanvasID);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not fetch Canvas grades.");
    }

    scoringInfos, err := assignment.GetScoringInfo(users);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to get scoring information.");
    }

    err = assignment.LatePolicy.Apply(assignment, scoringInfos, args.DryRun);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to apply late policy.");
    }

    err = computeFinalScores(users, scoringInfos, canvasGrades, args.DryRun);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to apply late policy.");
    }

    if (args.DryRun) {
        fmt.Println("Dry Run: Showing scoring infos instead of uploading them.");
        // TEST
        // fmt.Println(util.MustToJSONIndent(scoringInfos));
    }

    // TEST
    // fmt.Println("Uploaded assignment grades.");
}

// TEST
const LOCK_COMMENT string = "__lock__";

// TEST
func computeFinalScores(users map[string]*model.User, scoringInfos map[string]*model.ScoringInfo, canvasGrades []*canvas.CanvasGradeInfo, dryRun bool) error {
    var err error;

    // First, through canvas comments for locks and autograder notes.
    locks := make(map[string]bool);
    existingComments := make(map[string]*model.ScoringInfo);

    for _, canvasGradeInfo := range canvasGrades {
        for _, comment := range canvasGradeInfo.Comments {
            text := strings.ToLower(comment.Text);
            if (strings.Contains(text, LOCK_COMMENT)) {
                locks[canvasGradeInfo.UserID] = true;
            } else if (strings.Contains(text, model.SCORING_INFO_IDENTITY_KEY)) {
                var scoringInfo model.ScoringInfo;
                err = util.JSONFromString(comment.Text, &scoringInfo);
                if (err != nil) {
                    return fmt.Errorf("Could not unmarshall Canvas comment %s (%s) into a scoring info: '%w'.", comment.ID, comment.Text, err);
                }
                scoringInfo.CanvasCommentID = comment.ID;

                existingComments[canvasGradeInfo.UserID] = &scoringInfo;

                // Scoring infos can also lock grades.
                if (scoringInfo.Lock) {
                    locks[canvasGradeInfo.UserID] = true;
                }
            }
        }
    }

    // TEST - New comments can get added with the assignment. Old comments need to get updated.

    // Next, create the grades that will actually be uploaded.
    finalGrades := make([]*canvas.CanvasGradeInfo, 0);
    commentsToUpdate := make([]*model.ScoringInfo, 0);

    for email, scoringInfo := range scoringInfos {
        user := users[email];

        // This scoring is invalid, skip it.
        if (scoringInfo.Reject) {
            continue;
        }

        // Skip users that do not have a canvas id.
        if (user.CanvasID == "") {
            log.Warn().Str("user", email).Msg("User does not have a Canvas ID, skipping grade upload.");
            continue;
        }

        // Check the existing comment last so we can decide if this comment needs to be updated.
        existingComment := existingComments[user.CanvasID];
        if (existingComment != nil) {
            // If this user has an existing, then we may skip this upload if submission IDs match.
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
            commentsToUpdate = append(commentsToUpdate, scoringInfo);
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

    // TEST
    fmt.Println("TEST");
    fmt.Println("------- Grades --------")
    fmt.Println(util.MustToJSONIndent(finalGrades));
    fmt.Println("------- Comments --------")
    fmt.Println(util.MustToJSONIndent(commentsToUpdate));
    fmt.Println("-----------------------")

    // TEST
    return nil;
}

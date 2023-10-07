package model

import (
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/util"
)

type LateGradingPolicyType string;

const (
    // Apply no late policy at all.
    EmptyPolicy         LateGradingPolicyType = ""
    // Check the baseline (rejection), but nothing else.
    BaselinePolicy            LateGradingPolicyType = "baseline"
    ConstantPenalty     LateGradingPolicyType = "constant-penalty"
    PercentagePenalty   LateGradingPolicyType = "percentage-penalty"
    LateDays            LateGradingPolicyType = "late-days"
)

const (
    LATE_OPTIONS_KEY_PENALTY string = "penalty";
)

type LateGradingPolicy struct {
    Type LateGradingPolicyType `json:"type"`
    Penalty float64 `json:"penalty"`
    RejectAfterDays int `json:"reject-after-days"`

    MaxLateDays int `json:"max-late-days"`
    LateDaysCanvasID string `json:"late-days-canvas-id"`
}

type LateDaysInfo struct {
    AvailableDays int `json:"available-days"`
    UploadTime time.Time `json:"upload-time"`
    AllocatedDays map[string]int `json:"allocated-days"`

    // A distinct key so we can recognize this as an autograder object.
    Autograder int `json:"__autograder__v01__"`
    // If this object was serialized from a Canvas comment, keep the ID.
    CanvasCommentID string `json:"-"`
    CanvasCommentAuthorID string `json:"-"`
}

func (this *LateGradingPolicy) Validate() error {
    this.Type = LateGradingPolicyType(strings.ToLower(string(this.Type)));

    if (this.RejectAfterDays < 0) {
        return fmt.Errorf("Number of days for rejection is negative (%d), should be zero to be ignored or positive to be applied.", this.RejectAfterDays);
    }

    switch this.Type {
        case EmptyPolicy, BaselinePolicy:
            return nil;
        case ConstantPenalty:
            if (this.Penalty <= 0.0) {
                return fmt.Errorf("Policy '%s': penalty must be larger than zero, found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }
        case PercentagePenalty:
            if ((this.Penalty <= 0.0) || (this.Penalty > 1.0)) {
                return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }
        case LateDays:
            if ((this.Penalty <= 0.0) || (this.Penalty > 1.0)) {
                return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }

            if ((this.MaxLateDays < 1) || (this.MaxLateDays > this.RejectAfterDays)) {
                return fmt.Errorf("Policy '%s': max late days must be in [1, <reject days>(%d)], found '%d'.", this.Type, this.RejectAfterDays, this.MaxLateDays);
            }

            if (this.LateDaysCanvasID == "") {
                return fmt.Errorf("Policy '%s': Canvas ID for late days assignment cannot be empty.", this.Type);
            }
        default:
            return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
    }

    return nil;
}

// TEST - Need to update assignment comment.

// This assumes that all assignments are in canvas.
func (this *LateGradingPolicy) Apply(
        assignment *Assignment,
        users map[string]*User,
        scores map[string]*ScoringInfo,
        dryRun bool) error {
    // Start with each submission getting the raw score.
    for _, score := range scores {
        score.Score = score.RawScore;
    }

    // Empty policy does nothing.
    if (this.Type == EmptyPolicy) {
        return nil;
    }

    canvasAssignment, err := canvas.FetchAssignment(assignment.Course.CanvasInstanceInfo, assignment.CanvasID);
    if (err != nil) {
        return err;
    }

    dueDate, err := time.Parse(time.RFC3339, canvasAssignment.DueDate);
    if (err != nil) {
        return fmt.Errorf("Failed to parse canvas due date '%s': '%w'.", canvasAssignment.DueDate, err);
    }

    this.applyBaselinePolicy(users, scores, dueDate);

    // Baseline policy is complete.
    if (this.Type == BaselinePolicy) {
        return nil;
    }

    if ((this.Type == ConstantPenalty) || (this.Type == PercentagePenalty)) {
        penalty := this.Penalty;
        if (this.Type == PercentagePenalty) {
            penalty = canvasAssignment.MaxPoints * this.Penalty;
        }

        this.applyConstantPolicy(scores, penalty);
        return nil;
    }

    if (this.Type == LateDays) {
        penalty := canvasAssignment.MaxPoints * this.Penalty;
        err = this.applyLateDaysPolicy(assignment, users, scores, penalty, dryRun);
        if (err != nil) {
            return fmt.Errorf("Failed to apply late days policy: '%w'.", err);
        }

        return nil;
    }

    return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
}

// Apply a common policy.
func (this *LateGradingPolicy) applyBaselinePolicy(users map[string]*User, scores map[string]*ScoringInfo, dueDate time.Time) {
    for email, score := range scores {
        score.NumDaysLate = computeLateDays(dueDate, score.SubmissionTime);

        _, ok := users[email];
        if (!ok) {
            log.Warn().Str("user", email).Msg("Cannot find user, rejecting submission and skipping application of late polict.");
            score.Reject = true;
            continue;
        }

        if ((this.RejectAfterDays > 0) && (score.NumDaysLate > this.RejectAfterDays)) {
            score.Reject = true;
            continue;
        }
    }
}

// Apply a constant penalty per late day.
func (this *LateGradingPolicy) applyConstantPolicy(scores map[string]*ScoringInfo, penalty float64) {
    for _, score := range scores {
        if (score.NumDaysLate <= 0) {
            continue;
        }

        score.Score = math.Max(0.0, score.RawScore - (penalty * float64(score.NumDaysLate)));
    }
}

func (this *LateGradingPolicy) applyLateDaysPolicy(
        assignment *Assignment, users map[string]*User,
        scores map[string]*ScoringInfo, penalty float64,
        dryRun bool) error {
    allLateDays, err := this.fetchLateDays(assignment);
    if (err != nil) {
        return err;
    }

    lateDaysToUpdate := make(map[string]*LateDaysInfo);

    // TEST
    fmt.Println("%%%");
    // fmt.Println(util.MustToJSONIndent(allLateDays));
    fmt.Println("%%%");

    for email, scoringInfo := range scores {
        if (scoringInfo.Reject) {
            continue;
        }

        studentCanvasID := users[email].CanvasID;
        if (studentCanvasID == "") {
            log.Warn().Str("user", email).Msg("User does not have Canvas ID, cannot appply late days policy. Rejecting submission.");
            scoringInfo.Reject = true;
            continue;
        }

        lateDays := allLateDays[studentCanvasID];
        if (lateDays == nil) {
            log.Warn().Str("user", email).Str("canvas-id", studentCanvasID).Msg(
                    "Cannot find user late days, cannot appply late days policy. Rejecting submission.");
            scoringInfo.Reject = true;
            continue;
        }

        // Compute how many late days can be used.
        // To do this, we will reclaim any late days that have already been used in addition to free days.
        lateDaysAvailable := lateDays.AvailableDays;

        allocatedDays, hasAllocatedLateDays := lateDays.AllocatedDays[assignment.ID];
        if (hasAllocatedLateDays) {
            delete(lateDays.AllocatedDays, assignment.ID);
            lateDaysAvailable += allocatedDays;
        }

        // Assignment is not late and there are no records of allocating late days for this assignment, skip.
        // Late days could have been allocated if a future submission has been deleted.
        // We specifically did this after the earlier checks because we want to warn about the user and reject the submission.
        if ((scoringInfo.NumDaysLate <= 0) && !hasAllocatedLateDays) {
            continue;
        }

        // We will use late days limited by:
        // - The number of late days the user has to use.
        // - The maximum number of late days that can be used on this assignment.
        // - The number of days late the submission actually is.
        lateDaysToUse := min(lateDaysAvailable, this.MaxLateDays, scoringInfo.NumDaysLate);
        scoringInfo.LateDayUsage = lateDaysToUse;

        // Enforce a penalty for any remaining late days.
        remainingDaysLate := scoringInfo.NumDaysLate - lateDaysToUse;
        scoringInfo.Score = math.Max(0.0, scoringInfo.RawScore - (penalty * float64(remainingDaysLate)));

        // Check if the number of allocated late days has changed.
        // If so, we need to update the late days in canvas.
        if (allocatedDays != lateDaysToUse) {
            lateDays.AvailableDays = lateDaysAvailable - lateDaysToUse;
            lateDays.AllocatedDays[assignment.ID] = lateDaysToUse;
            lateDays.UploadTime = time.Now();

            lateDaysToUpdate[studentCanvasID] = lateDays;
        }
    }

    err = this.updateLateDays(assignment, lateDaysToUpdate, dryRun);
    if (err != nil) {
        return err;
    }

    return nil;
}

func (this *LateGradingPolicy) updateLateDays(assignment *Assignment, lateDaysToUpdate map[string]*LateDaysInfo, dryRun bool) error {
    // TEST
    fmt.Println("%%%");
    fmt.Println(util.MustToJSONIndent(lateDaysToUpdate));
    fmt.Println("%%%");

    // Update late days.
    // Info that does NOT have a CanvasCommentID will get the autograder comment added in.
    grades := make([]*canvas.CanvasGradeInfo, 0, len(lateDaysToUpdate));
    for canvasUser, lateInfo := range lateDaysToUpdate {
        uploadComments := make([]canvas.CanvasSubmissionComment, 0);
        if (lateInfo.CanvasCommentID == "") {
            uploadComments = append(uploadComments, canvas.CanvasSubmissionComment{
                Text: util.MustToJSON(lateInfo),
            });
        }

        gradeInfo := canvas.CanvasGradeInfo{
            UserID: canvasUser,
            Score: float64(lateInfo.AvailableDays),
            Time: lateInfo.UploadTime,
            Comments: uploadComments,
        }

        grades = append(grades, &gradeInfo);
    }

    if (dryRun) {
        // TEST
        // log.Info().Str("assignment", assignment.ID).Str("grades", util.MustToJSON(grades)).Msg("Dry Run: Skipping upload of late days.");
        log.Info().Str("assignment", assignment.ID).Any("grades", grades).Msg("Dry Run: Skipping upload of late days.");
    } else {
        err := canvas.UpdateAssignmentGrades(assignment.Course.CanvasInstanceInfo, assignment.CanvasID, grades);
        if (err != nil) {
            return fmt.Errorf("Failed to upload late days: '%w'.", err);
        }
    }

    // Update late days comment for info that has a CanvasCommentID.
    comments := make([]*canvas.CanvasSubmissionComment, 0, len(lateDaysToUpdate));
    for _, lateInfo := range lateDaysToUpdate {
        if (lateInfo.CanvasCommentID == "") {
            continue;
        }

        comments = append(comments, &canvas.CanvasSubmissionComment{
            ID: lateInfo.CanvasCommentID,
            Author: lateInfo.CanvasCommentAuthorID,
            Text: util.MustToJSON(lateInfo),
        });
    }

    if (dryRun) {
        log.Info().Str("assignment", assignment.ID).Any("comments", comments).Msg("Dry Run: Skipping update of late day comments.");
    } else {
        err := canvas.UpdateComments(assignment.Course.CanvasInstanceInfo, assignment.CanvasID, comments);
        if (err != nil) {
            return fmt.Errorf("Failed to update late days comments: '%w'.", err);
        }
    }

    return nil;
}

func (this *LateGradingPolicy) fetchLateDays(assignment *Assignment) (map[string]*LateDaysInfo, error) {
    // Fetch available late days from canvas.
    canvasLateDays, err := canvas.FetchAssignmentGrades(assignment.Course.CanvasInstanceInfo, this.LateDaysCanvasID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch late days assignment (%s): '%w'.", this.LateDaysCanvasID, err);
    }

    lateDays := make(map[string]*LateDaysInfo);

    // Parse out the full late days information.
    for _, canvasLateDayInfo := range canvasLateDays {
        var info LateDaysInfo;
        foundComment := false;

        // First check the comments for already submitted info.
        for _, comment := range canvasLateDayInfo.Comments {
            text := strings.ToLower(comment.Text);
            if (strings.Contains(text, canvas.LOCK_COMMENT)) {
                return nil, fmt.Errorf(
                        "Late days assignment '%s' for user '%s' has a lock comment. Resolve this lock to allow for grading.",
                        this.LateDaysCanvasID, canvasLateDayInfo.UserID);
            } else if (strings.Contains(text, AUTOGRADER_COMMENT_IDENTITY_KEY)) {
                err = util.JSONFromString(comment.Text, &info);
                if (err != nil) {
                    return nil, fmt.Errorf(
                            "Could not unmarshall Canvas comment %s (%s) into a late days info: '%w'.",
                            comment.ID, comment.Text, err);
                }
                info.CanvasCommentID = comment.ID;
                info.CanvasCommentAuthorID = comment.Author;

                foundComment = true;
            }
        }

        postedLateDays := int(math.Round(canvasLateDayInfo.Score));

        if (foundComment) {
            if (info.AvailableDays != postedLateDays) {
                log.Warn().Int("posted-days", postedLateDays).Int("comment-days", info.AvailableDays).Msg("Mismatch in the posted late days and the number found in the autograder comment.");
            }
        } else {
            info.AllocatedDays = make(map[string]int);
        }

        info.AvailableDays = postedLateDays;
        info.UploadTime = time.Now();

        lateDays[canvasLateDayInfo.UserID] = &info;
    }

    return lateDays, nil;
}

func computeLateDays(dueDate time.Time, submissionTime time.Time) int {
    if (dueDate.After(submissionTime)) {
        return 0;
    }

    return int(math.Ceil(submissionTime.Sub(dueDate).Hours() / 24.0));
}

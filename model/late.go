package model

import (
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
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
    LateDaysLMSID string `json:"late-days-lms-id"`
}

type LateDaysInfo struct {
    AvailableDays int `json:"available-days"`
    UploadTime time.Time `json:"upload-time"`
    AllocatedDays map[string]int `json:"allocated-days"`

    // A distinct key so we can recognize this as an autograder object.
    Autograder int `json:"__autograder__v01__"`
    // If this object was serialized from an LMS comment, keep the ID.
    LMSCommentID string `json:"-"`
    LMSCommentAuthorID string `json:"-"`
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

            if (this.LateDaysLMSID == "") {
                return fmt.Errorf("Policy '%s': LMS ID for late days assignment cannot be empty.", this.Type);
            }
        default:
            return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
    }

    return nil;
}

// This assumes that all assignments are in the LMS.
func (this *LateGradingPolicy) Apply(
        assignment *Assignment,
        users map[string]*usr.User,
        scores map[string]*artifact.ScoringInfo,
        dryRun bool) error {
    // Start with each submission getting the raw score.
    for _, score := range scores {
        score.Score = score.RawScore;
    }

    // Empty policy does nothing.
    if (this.Type == EmptyPolicy) {
        return nil;
    }

    lmsAssignment, err := assignment.GetCourse().GetLMSAdapter().FetchAssignment(assignment.GetLMSID());
    if (err != nil) {
        return err;
    }

    if (lmsAssignment.DueDate == nil) {
        return fmt.Errorf("Assignment does not have a due date.");
    }

    this.applyBaselinePolicy(users, scores, *lmsAssignment.DueDate);

    // Baseline policy is complete.
    if (this.Type == BaselinePolicy) {
        return nil;
    }

    if ((this.Type == ConstantPenalty) || (this.Type == PercentagePenalty)) {
        penalty := this.Penalty;
        if (this.Type == PercentagePenalty) {
            penalty = lmsAssignment.MaxPoints * this.Penalty;
        }

        this.applyConstantPolicy(scores, penalty);
        return nil;
    }

    if (this.Type == LateDays) {
        penalty := lmsAssignment.MaxPoints * this.Penalty;
        err = this.applyLateDaysPolicy(assignment, users, scores, penalty, dryRun);
        if (err != nil) {
            return fmt.Errorf("Failed to apply late days policy: '%w'.", err);
        }

        return nil;
    }

    return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
}

// Apply a common policy.
func (this *LateGradingPolicy) applyBaselinePolicy(users map[string]*usr.User, scores map[string]*artifact.ScoringInfo, dueDate time.Time) {
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
func (this *LateGradingPolicy) applyConstantPolicy(scores map[string]*artifact.ScoringInfo, penalty float64) {
    for _, score := range scores {
        if (score.NumDaysLate <= 0) {
            continue;
        }

        score.Score = math.Max(0.0, score.RawScore - (penalty * float64(score.NumDaysLate)));
    }
}

func (this *LateGradingPolicy) applyLateDaysPolicy(
        assignment *Assignment, users map[string]*usr.User,
        scores map[string]*artifact.ScoringInfo, penalty float64,
        dryRun bool) error {
    allLateDays, err := this.fetchLateDays(assignment);
    if (err != nil) {
        return err;
    }

    lateDaysToUpdate := make(map[string]*LateDaysInfo);

    for email, scoringInfo := range scores {
        if (scoringInfo.Reject) {
            continue;
        }

        studentLMSID := users[email].LMSID;
        if (studentLMSID == "") {
            log.Warn().Str("user", email).Msg("User does not have am LMS ID, cannot appply late days policy. Rejecting submission.");
            scoringInfo.Reject = true;
            continue;
        }

        lateDays := allLateDays[studentLMSID];
        if (lateDays == nil) {
            log.Warn().Str("user", email).Str("lms-id", studentLMSID).Msg(
                    "Cannot find user late days, cannot appply late days policy. Rejecting submission.");
            scoringInfo.Reject = true;
            continue;
        }

        // Compute how many late days can be used.
        // To do this, we will reclaim any late days that have already been used in addition to free days.
        lateDaysAvailable := lateDays.AvailableDays;

        allocatedDays, hasAllocatedLateDays := lateDays.AllocatedDays[assignment.GetID()];
        if (hasAllocatedLateDays) {
            delete(lateDays.AllocatedDays, assignment.GetID());
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
        // If so, we need to update the late days in the LMS.
        if (allocatedDays != lateDaysToUse) {
            lateDays.AvailableDays = lateDaysAvailable - lateDaysToUse;
            lateDays.AllocatedDays[assignment.GetID()] = lateDaysToUse;
            lateDays.UploadTime = time.Now();

            lateDaysToUpdate[studentLMSID] = lateDays;
        }
    }

    err = this.updateLateDays(assignment, lateDaysToUpdate, dryRun);
    if (err != nil) {
        return err;
    }

    return nil;
}

func (this *LateGradingPolicy) updateLateDays(assignment *Assignment, lateDaysToUpdate map[string]*LateDaysInfo, dryRun bool) error {
    // Update late days.
    // Info that does NOT have a LMSCommentID will get the autograder comment added in.
    grades := make([]*lms.SubmissionScore, 0, len(lateDaysToUpdate));
    for lmsUserID, lateInfo := range lateDaysToUpdate {
        uploadComments := make([]*lms.SubmissionComment, 0);
        if (lateInfo.LMSCommentID == "") {
            uploadComments = append(uploadComments, &lms.SubmissionComment{
                Text: util.MustToJSON(lateInfo),
            });
        }

        gradeInfo := lms.SubmissionScore{
            UserID: lmsUserID,
            Score: float64(lateInfo.AvailableDays),
            Time: lateInfo.UploadTime,
            Comments: uploadComments,
        }

        grades = append(grades, &gradeInfo);
    }

    if (dryRun) {
        log.Info().Str("assignment", assignment.GetID()).Any("grades", grades).Msg("Dry Run: Skipping upload of late days.");
    } else {
        err := assignment.GetCourse().GetLMSAdapter().UpdateAssignmentScores(this.LateDaysLMSID, grades);
        if (err != nil) {
            return fmt.Errorf("Failed to upload late days: '%w'.", err);
        }
    }

    // Update late days comment for info that has a LMSCommentID.
    comments := make([]*lms.SubmissionComment, 0, len(lateDaysToUpdate));
    for _, lateInfo := range lateDaysToUpdate {
        if (lateInfo.LMSCommentID == "") {
            continue;
        }

        comments = append(comments, &lms.SubmissionComment{
            ID: lateInfo.LMSCommentID,
            Author: lateInfo.LMSCommentAuthorID,
            Text: util.MustToJSON(lateInfo),
        });
    }

    if (dryRun) {
        log.Info().Str("assignment", assignment.GetID()).Any("comments", comments).Msg("Dry Run: Skipping update of late day comments.");
    } else {
        err := assignment.GetCourse().GetLMSAdapter().UpdateComments(this.LateDaysLMSID, comments);
        if (err != nil) {
            return fmt.Errorf("Failed to update late days comments: '%w'.", err);
        }
    }

    return nil;
}

func (this *LateGradingPolicy) fetchLateDays(assignment *Assignment) (map[string]*LateDaysInfo, error) {
    // Fetch available late days from the LMS.
    lmsLateDaysScores, err := assignment.GetCourse().GetLMSAdapter().FetchAssignmentScores(this.LateDaysLMSID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch late days assignment (%s): '%w'.", this.LateDaysLMSID, err);
    }

    lateDays := make(map[string]*LateDaysInfo);

    // Parse out the full late days information.
    for _, lmsLateDaysScore := range lmsLateDaysScores {
        var info LateDaysInfo;
        foundComment := false;

        // First check the comments for already submitted info.
        for _, comment := range lmsLateDaysScore.Comments {
            text := strings.ToLower(comment.Text);
            if (strings.Contains(text, LOCK_COMMENT)) {
                return nil, fmt.Errorf(
                        "Late days assignment '%s' for user '%s' has a lock comment. Resolve this lock to allow for grading.",
                        this.LateDaysLMSID, lmsLateDaysScore.UserID);
            } else if (strings.Contains(text, common.AUTOGRADER_COMMENT_IDENTITY_KEY)) {
                err = util.JSONFromString(comment.Text, &info);
                if (err != nil) {
                    return nil, fmt.Errorf(
                            "Could not unmarshall LSM comment %s (%s) into a late days info: '%w'.",
                            comment.ID, comment.Text, err);
                }
                info.LMSCommentID = comment.ID;
                info.LMSCommentAuthorID = comment.Author;

                foundComment = true;
            }
        }

        postedLateDays := int(math.Round(lmsLateDaysScore.Score));

        if (foundComment) {
            if (info.AvailableDays != postedLateDays) {
                log.Warn().Int("posted-days", postedLateDays).Int("comment-days", info.AvailableDays).Msg("Mismatch in the posted late days and the number found in the autograder comment.");
            }
        } else {
            info.AllocatedDays = make(map[string]int);
        }

        info.AvailableDays = postedLateDays;
        info.UploadTime = time.Now();

        lateDays[lmsLateDaysScore.UserID] = &info;
    }

    return lateDays, nil;
}

func computeLateDays(dueDate time.Time, submissionTime time.Time) int {
    if (dueDate.After(submissionTime)) {
        return 0;
    }

    return int(math.Ceil(submissionTime.Sub(dueDate).Hours() / 24.0));
}

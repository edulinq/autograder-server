package scoring

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

const LATE_DAYS_STRUCT_VERSION = "1.0.0"

type LateDaysInfo struct {
	AvailableDays int                 `json:"available-days"`
	UploadTime    timestamp.Timestamp `json:"upload-time"`
	AllocatedDays map[string]int      `json:"allocated-days"`

	// Stores the penalty (points per late day) for each assignment.
	// Used for optimal reallocation when new assignments are graded.
	AllocationValues map[string]float64 `json:"allocation-values,omitempty"`

	// Stores the number of days late for each assignment.
	// This is needed to know the maximum late days that could be used per assignment.
	DaysLatePerAssignment map[string]int `json:"days-late,omitempty"`

	// Stores the submission time for each assignment.
	// Used for standard allocation ordering (earliest submission first).
	SubmissionTimes map[string]timestamp.Timestamp `json:"submission-times,omitempty"`

	// A distinct key so we can recognize this as an autograder object.
	AutograderStructVersion string `json:"__autograder__version__"`

	// If this object was serialized from an LMS comment, keep the ID.
	LMSCommentID       string `json:"-"`
	LMSCommentAuthorID string `json:"-"`
}

// This assumes that all assignments are in the LMS.
func ApplyLatePolicy(
	assignment *model.Assignment,
	users map[string]*model.CourseUser,
	scores map[string]*model.ScoringInfo,
	dryRun bool) error {
	policy := assignment.GetLatePolicy()

	// Start with each submission getting the raw score.
	for _, score := range scores {
		score.Score = score.RawScore
	}

	// Empty policy does nothing.
	if policy.Type == model.EmptyPolicy {
		return nil
	}

	lmsAssignment, err := lms.FetchAssignment(assignment.GetCourse(), assignment.GetLMSID())
	if err != nil {
		return err
	}

	if lmsAssignment.DueDate == nil {
		return fmt.Errorf("Assignment does not have a due date.")
	}

	applyBaselinePolicy(assignment, policy, users, scores, *lmsAssignment.DueDate)

	// Baseline policy is complete.
	if policy.Type == model.BaselinePolicy {
		return nil
	}

	if (policy.Type == model.ConstantPenalty) || (policy.Type == model.PercentagePenalty) {
		penalty := policy.Penalty
		if policy.Type == model.PercentagePenalty {
			penalty = lmsAssignment.MaxPoints * policy.Penalty
		}

		applyConstantPolicy(policy, scores, penalty)
		return nil
	}

	if policy.Type == model.LateDays {
		penalty := lmsAssignment.MaxPoints * policy.Penalty
		err = applyLateDaysPolicy(policy, assignment, users, scores, penalty, dryRun)
		if err != nil {
			return fmt.Errorf("Failed to apply late days policy: '%w'.", err)
		}

		return nil
	}

	return fmt.Errorf("Unknown late policy type: '%s'.", policy.Type)
}

// Apply a common policy.
func applyBaselinePolicy(assignment *model.Assignment, policy model.LateGradingPolicy, users map[string]*model.CourseUser, scores map[string]*model.ScoringInfo, dueDate timestamp.Timestamp) {
	for email, score := range scores {
		score.NumDaysLate = computeLateDays(dueDate, score.SubmissionTime, policy.GraceMinutes)

		_, ok := users[email]
		if !ok {
			log.Warn("Cannot find user, rejecting submission and skipping application of late polict.", assignment, log.NewUserAttr(email))
			score.Reject = true
			continue
		}

		if (policy.RejectAfterDays > 0) && (score.NumDaysLate > policy.RejectAfterDays) {
			score.Reject = true
			continue
		}
	}
}

// Apply a constant penalty per late day.
func applyConstantPolicy(policy model.LateGradingPolicy, scores map[string]*model.ScoringInfo, penalty float64) {
	for _, score := range scores {
		if score.NumDaysLate <= 0 {
			continue
		}

		score.Score = math.Max(0.0, score.RawScore-(penalty*float64(score.NumDaysLate)))
	}
}

func applyLateDaysPolicy(policy model.LateGradingPolicy, assignment *model.Assignment, users map[string]*model.CourseUser, scores map[string]*model.ScoringInfo, penalty float64, dryRun bool) error {
	if policy.LateDaysLMSID == "" {
		return fmt.Errorf("Cannot apply late days policy, late days assignment LMS ID is empty.")
	}

	allLateDays, err := fetchLateDays(policy, assignment)
	if err != nil {
		return err
	}

	// Get optimal allocation from course-level late policy (not per-assignment).
	// This ensures consistent behavior across all assignments in the course.
	optimalAllocation := false
	if assignment.GetCourse().LatePolicy != nil {
		optimalAllocation = assignment.GetCourse().LatePolicy.OptimalAllocation
	}

	lateDaysToUpdate := make(map[string]*LateDaysInfo)

	for email, scoringInfo := range scores {
		if scoringInfo.Reject {
			continue
		}

		studentLMSID := users[email].GetLMSID()
		if studentLMSID == "" {
			log.Warn("User does not have am LMS ID, cannot appply late days policy. Rejecting submission.",
				assignment, users[email])
			scoringInfo.Reject = true
			continue
		}

		lateDays := allLateDays[studentLMSID]
		if lateDays == nil {
			log.Warn("Cannot find user late days, cannot appply late days policy. Rejecting submission.",
				assignment, users[email], log.NewAttr("lms-id", studentLMSID))
			scoringInfo.Reject = true
			continue
		}

		// Initialize maps if they don't exist (for backwards compatibility).
		if lateDays.AllocationValues == nil {
			lateDays.AllocationValues = make(map[string]float64)
		}

		if lateDays.DaysLatePerAssignment == nil {
			lateDays.DaysLatePerAssignment = make(map[string]int)
		}

		if lateDays.SubmissionTimes == nil {
			lateDays.SubmissionTimes = make(map[string]timestamp.Timestamp)
		}

		// Save original allocations to detect any changes (not just current assignment).
		// This is needed because optimal allocation may reallocate late days between assignments.
		originalAllocations := make(map[string]int)
		for k, v := range lateDays.AllocatedDays {
			originalAllocations[k] = v
		}

		// Compute how many late days can be used.
		// To do this, we will reclaim any late days that have already been used in addition to free days.
		lateDaysAvailable := lateDays.AvailableDays

		allocatedDays, hasAllocatedLateDays := lateDays.AllocatedDays[assignment.GetID()]
		if hasAllocatedLateDays {
			delete(lateDays.AllocatedDays, assignment.GetID())
			lateDaysAvailable += allocatedDays
		}

		// Assignment is not late and there are no records of allocating late days for this assignment, skip.
		// Late days could have been allocated if a future submission has been deleted.
		// We specifically did this after the earlier checks because we want to warn about the user and reject the submission.
		if (scoringInfo.NumDaysLate <= 0) && !hasAllocatedLateDays {
			continue
		}

		// Use unified allocation function for both standard and optimal modes.
		// Standard mode: allocates by submission time (earliest first).
		// Optimal mode: allocates by penalty value (highest points saved first).
		lateDaysToUse := computeLateDayAllocation(
			lateDays,
			assignment.GetID(),
			scoringInfo.NumDaysLate,
			penalty,
			scoringInfo.SubmissionTime,
			policy.MaxLateDays,
			lateDaysAvailable,
			optimalAllocation,
		)

		scoringInfo.LateDayUsage = lateDaysToUse

		// Enforce a penalty for any remaining late days.
		remainingDaysLate := scoringInfo.NumDaysLate - lateDaysToUse
		scoringInfo.Score = math.Max(0.0, scoringInfo.RawScore-(penalty*float64(remainingDaysLate)))

		// Update current assignment's allocation.
		lateDays.AllocatedDays[assignment.GetID()] = lateDaysToUse

		// Check if any allocation changed (not just the current assignment).
		// This handles the case where grading assignment A causes reallocation from B to C.
		allocationsChanged := !mapsEqual(originalAllocations, lateDays.AllocatedDays)

		if allocationsChanged {
			lateDays.UploadTime = timestamp.Now()

			// Store allocation metadata for future reallocation.
			lateDays.AllocationValues[assignment.GetID()] = penalty
			lateDays.DaysLatePerAssignment[assignment.GetID()] = scoringInfo.NumDaysLate
			lateDays.SubmissionTimes[assignment.GetID()] = scoringInfo.SubmissionTime

			lateDaysToUpdate[studentLMSID] = lateDays
		}
	}

	err = updateLateDays(policy, assignment, lateDaysToUpdate, dryRun)
	if err != nil {
		return err
	}

	return nil
}

func updateLateDays(policy model.LateGradingPolicy, assignment *model.Assignment, lateDaysToUpdate map[string]*LateDaysInfo, dryRun bool) error {
	// Update late days.
	// Info that does NOT have a LMSCommentID will get the autograder comment added in.
	grades := make([]*lmstypes.SubmissionScore, 0, len(lateDaysToUpdate))
	for lmsUserID, lateInfo := range lateDaysToUpdate {
		uploadComments := make([]*lmstypes.SubmissionComment, 0)
		if lateInfo.LMSCommentID == "" {
			uploadComments = append(uploadComments, &lmstypes.SubmissionComment{
				Text: util.MustToJSON(lateInfo),
			})
		}

		gradeInfo := lmstypes.SubmissionScore{
			UserID:   lmsUserID,
			Score:    float64(lateInfo.AvailableDays),
			Time:     &lateInfo.UploadTime,
			Comments: uploadComments,
		}

		grades = append(grades, &gradeInfo)
	}

	if dryRun {
		log.Debug("Dry Run: Skipping upload of late days.", assignment, log.NewAttr("grades", grades))
	} else {
		err := lms.UpdateAssignmentScores(assignment.GetCourse(), policy.LateDaysLMSID, grades)
		if err != nil {
			return fmt.Errorf("Failed to upload late days: '%w'.", err)
		}
	}

	// Update late days comment for info that has a LMSCommentID.
	comments := make([]*lmstypes.SubmissionComment, 0, len(lateDaysToUpdate))
	for _, lateInfo := range lateDaysToUpdate {
		if lateInfo.LMSCommentID == "" {
			continue
		}

		comments = append(comments, &lmstypes.SubmissionComment{
			ID:     lateInfo.LMSCommentID,
			Author: lateInfo.LMSCommentAuthorID,
			Text:   util.MustToJSON(lateInfo),
		})
	}

	if dryRun {
		log.Debug("Dry Run: Skipping update of late day comments.", assignment, log.NewAttr("comments", comments))
	} else {
		err := lms.UpdateComments(assignment.GetCourse(), policy.LateDaysLMSID, comments)
		if err != nil {
			return fmt.Errorf("Failed to update late days comments: '%w'.", err)
		}
	}

	return nil
}

func fetchLateDays(policy model.LateGradingPolicy, assignment *model.Assignment) (map[string]*LateDaysInfo, error) {
	// Fetch available late days from the LMS.
	lmsLateDaysScores, err := lms.FetchAssignmentScores(assignment.GetCourse(), policy.LateDaysLMSID)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch late days assignment (%s): '%w'.", policy.LateDaysLMSID, err)
	}

	lateDays := make(map[string]*LateDaysInfo)

	// Parse out the full late days information.
	for _, lmsLateDaysScore := range lmsLateDaysScores {
		var info LateDaysInfo
		foundComment := false

		// First check the comments for already submitted info.
		for _, comment := range lmsLateDaysScore.Comments {
			text := strings.ToLower(comment.Text)
			if strings.Contains(text, LOCK_COMMENT) {
				return nil, fmt.Errorf(
					"Late days assignment '%s' for user '%s' has a lock comment. Resolve this lock to allow for grading.",
					policy.LateDaysLMSID, lmsLateDaysScore.UserID)
			} else if strings.Contains(text, common.AUTOGRADER_COMMENT_IDENTITY_KEY) {
				err = util.JSONFromString(comment.Text, &info)
				if err != nil {
					return nil, fmt.Errorf(
						"Could not unmarshall LSM comment %s (%s) into a late days info: '%w'.",
						comment.ID, comment.Text, err)
				}
				info.LMSCommentID = comment.ID
				info.LMSCommentAuthorID = comment.Author

				if LATE_DAYS_STRUCT_VERSION != info.AutograderStructVersion {
					return nil, fmt.Errorf("Mismatch in late days info version found in LMS comment. Current version: '%s', comment version: '%s'.",
						LATE_DAYS_STRUCT_VERSION, info.AutograderStructVersion)
				}

				foundComment = true
			}
		}

		postedLateDays := int(math.Round(lmsLateDaysScore.Score))

		if foundComment {
			if info.AvailableDays != postedLateDays {
				log.Warn("Mismatch in the posted late days and the number found in the autograder comment.",
					assignment, log.NewAttr("posted-days", postedLateDays), log.NewAttr("comment-days", info.AvailableDays), log.NewAttr("user-lms-id", lmsLateDaysScore.UserID))
			}
		} else {
			info.AllocatedDays = make(map[string]int)
		}

		info.AvailableDays = postedLateDays
		info.UploadTime = timestamp.Now()
		info.AutograderStructVersion = LATE_DAYS_STRUCT_VERSION

		lateDays[lmsLateDaysScore.UserID] = &info
	}

	return lateDays, nil
}

// assignmentAllocation represents a potential late day allocation for an assignment.
type assignmentAllocation struct {
	AssignmentID   string
	DaysLate       int
	PointsPerDay   float64             // Points saved per late day used
	SubmissionTime timestamp.Timestamp // When the assignment was submitted
	MaxDaysAllowed int                 // Maximum late days that can be used for this assignment
}

// computeLateDayAllocation allocates late days across all assignments using a greedy algorithm.
// In standard mode (optimalMode=false), assignments are processed by submission time (earliest first).
// In optimal mode (optimalMode=true), assignments are processed by penalty value (highest first).
//
// Note: This optimization is based on raw assignment scores only.
// External weighting (like exam vs quiz categories) is not considered.
func computeLateDayAllocation(
	lateDays *LateDaysInfo,
	currentAssignmentID string,
	currentDaysLate int,
	currentPenalty float64,
	currentSubmissionTime timestamp.Timestamp,
	maxLateDaysPerAssignment int,
	totalLateDaysAvailable int,
	optimalMode bool,
) int {
	// Build a list of all assignments that need late days.
	allocations := make([]assignmentAllocation, 0)

	// Add current assignment.
	allocations = append(allocations, assignmentAllocation{
		AssignmentID:   currentAssignmentID,
		DaysLate:       currentDaysLate,
		PointsPerDay:   currentPenalty,
		SubmissionTime: currentSubmissionTime,
		MaxDaysAllowed: min(maxLateDaysPerAssignment, currentDaysLate),
	})

	// Add previously allocated assignments.
	for assignmentID, daysLate := range lateDays.DaysLatePerAssignment {
		if assignmentID == currentAssignmentID {
			continue // Already added above.
		}

		pointsPerDay, hasValue := lateDays.AllocationValues[assignmentID]
		submissionTime, hasTime := lateDays.SubmissionTimes[assignmentID]
		if !hasValue || !hasTime || daysLate <= 0 {
			continue
		}

		// Reclaim previously allocated days for this assignment.
		if allocatedDays, ok := lateDays.AllocatedDays[assignmentID]; ok {
			totalLateDaysAvailable += allocatedDays
		}

		allocations = append(allocations, assignmentAllocation{
			AssignmentID:   assignmentID,
			DaysLate:       daysLate,
			PointsPerDay:   pointsPerDay,
			SubmissionTime: submissionTime,
			MaxDaysAllowed: min(maxLateDaysPerAssignment, daysLate),
		})
	}

	// Sort based on mode.
	if optimalMode {
		// Optimal mode: sort by points per day (descending) - highest value first.
		sort.Slice(allocations, func(i, j int) bool {
			return allocations[i].PointsPerDay > allocations[j].PointsPerDay
		})
	} else {
		// Standard mode: sort by submission time (ascending) - earliest first.
		sort.Slice(allocations, func(i, j int) bool {
			return allocations[i].SubmissionTime < allocations[j].SubmissionTime
		})
	}

	// Greedily allocate late days to assignments based on sort order.
	remainingLateDays := totalLateDaysAvailable
	newAllocations := make(map[string]int)

	for _, alloc := range allocations {
		if remainingLateDays <= 0 {
			break
		}

		daysToAllocate := min(remainingLateDays, alloc.MaxDaysAllowed)
		if daysToAllocate > 0 {
			newAllocations[alloc.AssignmentID] = daysToAllocate
			remainingLateDays -= daysToAllocate
		}
	}

	// Update the lateDays struct with new allocations for other assignments.
	for assignmentID, days := range newAllocations {
		if assignmentID != currentAssignmentID {
			lateDays.AllocatedDays[assignmentID] = days
		}
	}

	// Update available days (will be adjusted later for current assignment).
	lateDays.AvailableDays = remainingLateDays

	// Return the allocation for the current assignment.
	return newAllocations[currentAssignmentID]
}

func computeLateDays(dueDate timestamp.Timestamp, submissionTime timestamp.Timestamp, graceMinutes int) int {
	// Apply grace time to the due date.
	// Convert grace minutes to milliseconds (minutes * 60 seconds * 1000 msecs).
	graceMSecs := int64(graceMinutes) * 60 * 1000
	adjustedDueDate := dueDate + timestamp.Timestamp(graceMSecs)

	if adjustedDueDate >= submissionTime {
		return 0
	}

	delta := submissionTime.ToMSecs() - adjustedDueDate.ToMSecs()

	// Convert delta (msecs) to seconds -> minutes -> hours -> days.
	return int(math.Ceil(float64(delta) / 1000.0 / 60.0 / 60.0 / 24.0))
}

// mapsEqual checks if two maps have the same keys and values.
func mapsEqual(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}

	return true
}

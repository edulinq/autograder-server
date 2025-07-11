package scoring

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

const LOCK_COMMENT string = "__lock__"

func FullAssignmentScoringAndUpload(assignment *model.Assignment, dryRun bool) (map[string]*model.ScoringInfo, error) {
	if assignment.GetCourse().GetLMSAdapter() == nil {
		return nil, fmt.Errorf("Assignment's course has no LMS info associated with it.")
	}

	users, err := db.GetCourseUsers(assignment.GetCourse())
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch autograder users: '%w'.", err)
	}

	lmsScores, err := lms.FetchAssignmentScores(assignment.GetCourse(), assignment.GetLMSID())
	if err != nil {
		return nil, fmt.Errorf("Could not fetch LMS grades: '%w'.", err)
	}

	reference := model.CourseUserRoleToParsedCourseUserReference(model.CourseRoleStudent)

	scoringInfos, err := db.GetExistingScoringInfos(assignment, reference)
	if err != nil {
		return nil, fmt.Errorf("Failed to get scoring information: '%w'.", err)
	}

	err = ApplyLatePolicy(assignment, users, scoringInfos, dryRun)
	if err != nil {
		return nil, fmt.Errorf("Failed to apply late policy: '%w'.", err)
	}

	uploadedScores, err := computeFinalScores(assignment, users, scoringInfos, lmsScores, dryRun)
	if err != nil {
		return nil, fmt.Errorf("Failed to apply late policy: '%w'.", err)
	}

	return uploadedScores, nil
}

func computeFinalScores(
	assignment *model.Assignment, users map[string]*model.CourseUser, scoringInfos map[string]*model.ScoringInfo, lmsScores []*lmstypes.SubmissionScore, dryRun bool) (map[string]*model.ScoringInfo, error) {
	var err error

	// First, look through comments for locks and autograder notes.
	locks, existingComments, err := parseComments(lmsScores)
	if err != nil {
		return nil, err
	}

	// Next, create the grades that will actually be uploaded and the comments that will be updated..
	usedScoringInfos, finalScores, commentsToUpdate := filterFinalScores(assignment, users, scoringInfos, locks, existingComments)

	// Upload the grades.
	if dryRun {
		log.Debug("Dry Run: Skipping upload of final grades.", assignment, log.NewAttr("grades", finalScores))
	} else {
		err = lms.UpdateAssignmentScores(assignment.GetCourse(), assignment.GetLMSID(), finalScores)
		if err != nil {
			return nil, fmt.Errorf("Failed to upload final scores: '%w'.", err)
		}
	}

	// Update the comments.
	if dryRun {
		log.Debug("Dry Run: Skipping update of final comments.", assignment, log.NewAttr("comments", commentsToUpdate))
	} else {
		err = lms.UpdateComments(assignment.GetCourse(), assignment.GetLMSID(), commentsToUpdate)
		if err != nil {
			return nil, fmt.Errorf("Failed to update final comments: '%w'.", err)
		}
	}

	return usedScoringInfos, nil
}

func parseComments(lmsScores []*lmstypes.SubmissionScore) (map[string]bool, map[string]*model.ScoringInfo, error) {
	locks := make(map[string]bool)
	existingComments := make(map[string]*model.ScoringInfo)

	for _, lmsScore := range lmsScores {
		for _, comment := range lmsScore.Comments {
			text := strings.ToLower(comment.Text)
			if strings.Contains(text, LOCK_COMMENT) {
				locks[lmsScore.UserID] = true
			} else if strings.Contains(text, common.AUTOGRADER_COMMENT_IDENTITY_KEY) {
				var scoringInfo model.ScoringInfo
				err := util.JSONFromString(comment.Text, &scoringInfo)
				if err != nil {
					return nil, nil, fmt.Errorf("Could not unmarshall LMS comment %s (%s) into a scoring info: '%w'.", comment.ID, comment.Text, err)
				}

				if model.SCORING_INFO_STRUCT_VERSION != scoringInfo.AutograderStructVersion {
					return nil, nil, fmt.Errorf("Mismatch in late days info version found in LMS comment. Current version: '%s', comment version: '%s'.",
						model.SCORING_INFO_STRUCT_VERSION, scoringInfo.AutograderStructVersion)
				}

				scoringInfo.LMSCommentID = comment.ID
				scoringInfo.LMSCommentAuthorID = comment.Author

				existingComments[lmsScore.UserID] = &scoringInfo

				// Scoring infos can also lock grades.
				if scoringInfo.Lock {
					locks[lmsScore.UserID] = true
				}
			}
		}
	}

	return locks, existingComments, nil
}

func filterFinalScores(
	assignment *model.Assignment,
	users map[string]*model.CourseUser, scoringInfos map[string]*model.ScoringInfo,
	locks map[string]bool, existingComments map[string]*model.ScoringInfo,
) (map[string]*model.ScoringInfo, []*lmstypes.SubmissionScore, []*lmstypes.SubmissionComment) {
	usedScoringInfos := make(map[string]*model.ScoringInfo, 0)
	finalScores := make([]*lmstypes.SubmissionScore, 0)
	commentsToUpdate := make([]*lmstypes.SubmissionComment, 0)

	for email, scoringInfo := range scoringInfos {
		user := users[email]
		if user == nil {
			log.Warn("User does not exist, skipping grade upload.", assignment, log.NewUserAttr(email))
			continue
		}

		// This scoring is invalid, skip it.
		if scoringInfo.Reject {
			continue
		}

		// Skip users that do not have an LMS id.
		lmsID := user.GetLMSID()
		if lmsID == "" {
			log.Warn("User does not have an LMS ID, skipping grade upload.", assignment, log.NewUserAttr(email))
			continue
		}

		// This score is locked, skip it.
		if locks[lmsID] {
			continue
		}

		scoringInfo.UploadTime = timestamp.Now()

		// Check the existing comment last so we can decide if this comment needs to be updated.
		existingComment := existingComments[lmsID]
		if existingComment != nil {
			// If this user has an existing comment, then we may skip this upload if everything matches.
			if existingComment.Equal(scoringInfo) {
				log.Trace("User's submission/grade is up-to-date.",
					assignment, log.NewUserAttr(email), log.NewAttr("submittion-id", existingComment.ID))
				continue
			}
		}

		// This scoring is valid and different than the last one.

		// Existing comments are updated, new comments are posted with the grade.
		var uploadComments []*lmstypes.SubmissionComment = nil

		if existingComment != nil {
			scoringInfo.LMSCommentID = existingComment.LMSCommentID
			scoringInfo.LMSCommentAuthorID = existingComment.LMSCommentAuthorID

			commentsToUpdate = append(commentsToUpdate, &lmstypes.SubmissionComment{
				ID:     scoringInfo.LMSCommentID,
				Author: scoringInfo.LMSCommentAuthorID,
				Text:   util.MustToJSON(scoringInfo),
			})
		} else {
			uploadComments = []*lmstypes.SubmissionComment{
				&lmstypes.SubmissionComment{
					Text: util.MustToJSON(scoringInfo),
				},
			}
		}

		lmsScore := lmstypes.SubmissionScore{
			UserID:   lmsID,
			Score:    scoringInfo.Score,
			Time:     &scoringInfo.SubmissionTime,
			Comments: uploadComments,
		}

		finalScores = append(finalScores, &lmsScore)
		usedScoringInfos[email] = scoringInfo
	}

	return usedScoringInfos, finalScores, commentsToUpdate
}

package scores

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Change the course name and ensure it is back after an update.
func TestLMSScoresUpload(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Set the assignment's LMS ID.
	course := db.MustGetTestCourse()
	course.Assignments["hw0"].LMSID = "001"
	err := db.SaveCourse(course)
	if err != nil {
		test.Fatalf("Failed to save course: '%v'.", err)
	}

	response := core.SendTestAPIRequest(test, `courses/lms/scores/upload`, nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent UploadResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if len(responseContent.Results) != 1 {
		test.Fatalf("Incorrect number of results. Expected: 1, Actual: %d.", len(responseContent.Results))
	}

	// Zero out the upload time.
	responseContent.Results[0].UploadTime = timestamp.Zero()

	expectedResponse := UploadResponse{
		DryRun: false,
		Results: []*model.ExternalScoringInfo{
			&model.ExternalScoringInfo{
				UserEmail:      "course-student@test.edulinq.org",
				AssignmentID:   "hw0",
				SubmissionID:   "course101::hw0::course-student@test.edulinq.org::1697406272",
				SubmissionTime: timestamp.FromMSecs(1697406273000),
				UploadTime:     timestamp.Zero(),
				RawScore:       2,
				Score:          2,
			},
		},
	}

	if !reflect.DeepEqual(expectedResponse, responseContent) {
		test.Fatalf("Result not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expectedResponse), util.MustToJSONIndent(responseContent))
	}
}

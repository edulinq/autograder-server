package canvas

import (
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/util"
)

var testScore lmstypes.SubmissionScore = lmstypes.SubmissionScore{
	UserID: "00040",
	Score:  100.0,
	Time:   time.Time{},
	Comments: []*lmstypes.SubmissionComment{
		&lmstypes.SubmissionComment{
			ID:     "0987654",
			Author: "7827",
			Text:   "{\n\"id\": \"course101::hw0::student@test.edulinq.org::1696364768\",\n\"submission-time\": \"2023-10-03T20:26:08.951546Z\",\n\"upload-time\": \"2023-10-07T13:04:54.412979316-05:00\",\n\"raw-score\": 100,\n\"score\": 100,\n\"lock\": false,\n\"late-date-usage\": 0,\n\"num-days-late\": 0,\n\"reject\": false,\n\"__autograder__v01__\": 0\n}",
			Time:   "",
		},
	},
}

func TestFetchAssignmentSccoreBase(test *testing.T) {
	score, err := testBackend.FetchAssignmentScore(TEST_ASSIGNMENT_ID, "00040")
	if err != nil {
		test.Fatalf("Failed to fetch assignment score: '%v'.", err)
	}

	// Can't compare directly because of time.Time.
	// Use JSON instead.
	expectedJSON := util.MustToJSONIndent(testScore)
	actualJSON := util.MustToJSONIndent(score)

	if expectedJSON != actualJSON {
		test.Fatalf("Score not as expected. Expected: '%s', Actual: '%s'.",
			expectedJSON, actualJSON)
	}
}

func TestFetchAssignmentSccoresBase(test *testing.T) {
	scores, err := testBackend.fetchAssignmentScores(TEST_ASSIGNMENT_ID, true)
	if err != nil {
		test.Fatalf("Failed to fetch assignment scores: '%v'.", err)
	}

	expected := []*lmstypes.SubmissionScore{
		&testScore,
		&testScore,
		&testScore,
	}

	// Can't compare directly because of time.Time.
	// Use JSON instead.
	expectedJSON := util.MustToJSONIndent(expected)
	actualJSON := util.MustToJSONIndent(scores)

	if expectedJSON != actualJSON {
		test.Fatalf("Scores not as expected. Expected: '%s', Actual: '%s'.",
			expectedJSON, actualJSON)
	}
}

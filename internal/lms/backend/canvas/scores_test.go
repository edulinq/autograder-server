package canvas

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/util"
)

var testScore lmstypes.SubmissionScore = lmstypes.SubmissionScore{
	UserID: "00040",
	Score:  100.0,
	Time:   nil,
	Comments: []*lmstypes.SubmissionComment{
		&lmstypes.SubmissionComment{
			ID:     "0987654",
			Author: "7827",
			Text:   "{\n\"id\": \"course101::hw0::course-student@test.edulinq.org::1696364768\",\n\"submission-time\":1234,\n\"upload-time\":1235,\n\"raw-score\": 100,\n\"score\": 100,\n\"lock\": false,\n\"late-date-usage\": 0,\n\"num-days-late\": 0,\n\"reject\": false,\n\"__autograder__v01__\": 0\n}",
			Time:   "",
		},
	},
}

func TestFetchAssignmentScoreBase(test *testing.T) {
	score, err := testBackend.FetchAssignmentScore(TEST_ASSIGNMENT_ID, "00040")
	if err != nil {
		test.Fatalf("Failed to fetch assignment score: '%v'.", err)
	}

	if !reflect.DeepEqual(&testScore, score) {
		test.Fatalf("Score not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testScore), util.MustToJSONIndent(score))
	}
}

func TestFetchAssignmentScoresBase(test *testing.T) {
	scores, err := testBackend.fetchAssignmentScores(TEST_ASSIGNMENT_ID, true)
	if err != nil {
		test.Fatalf("Failed to fetch assignment scores: '%v'.", err)
	}

	expected := []*lmstypes.SubmissionScore{
		&testScore,
		&testScore,
		&testScore,
	}

	if !reflect.DeepEqual(expected, scores) {
		test.Fatalf("Scores not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(scores))
	}
}

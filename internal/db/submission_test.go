package db

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestRemoveSubmission(test *testing.T) {
	defer ResetForTesting()

	testCases := []struct {
		targetEmail      string
		targetSubmission string
		isRemoved        bool
	}{
		// Specific email, specific submission.
		{"course-student@test.edulinq.org", "1697406256", true},
		{"course-student@test.edulinq.org", "1697406265", true},
		{"course-student@test.edulinq.org", "1697406272", true},

		// Specific email, specific submission (full ID).
		{"course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406256", true},
		{"course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406265", true},
		{"course-student@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406272", true},

		// Specific email, missing submission.
		{"course-student@test.edulinq.org", "ZZZ", false},

		// Specific email, recent submission.
		{"course-student@test.edulinq.org", "", true},

		// Missing email, specific submission.
		{"ZZZ@test.edulinq.org", "1697406256", false},
		{"ZZZ@test.edulinq.org", "1697406265", false},
		{"ZZZ@test.edulinq.org", "1697406272", false},

		// Missing email, missing submission.
		{"ZZZ@test.edulinq.org", "ZZZ", false},

		// Missing email, specific submission (full ID).
		{"ZZZ@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406256", false},
		{"ZZZ@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406265", false},
		{"ZZZ@test.edulinq.org", "course101::hw0::student@test.edulinq.org::1697406272", false},

		// Missing email, recent submission.
		{"ZZZ@test.edulinq.org", "", false},
	}

	for i, testCase := range testCases {
		// Reload the test course every time.
		ResetForTesting()

		assignment := MustGetTestAssignment()

		isRemoved, err := RemoveSubmission(assignment, testCase.targetEmail, testCase.targetSubmission)
		if err != nil {
			test.Errorf("Case %d: Submission removal failed: '%v'.", i, err)
			continue
		}

		if isRemoved != testCase.isRemoved {
			test.Errorf("Case %d: Removed submission does not match. Expected : '%v', actual: '%v'.", i, testCase.isRemoved, isRemoved)
			continue
		}
	}
}

// Tests GetSubmissionAttempts as follows:
// A) Fetch all attempts from a user who has submissions and check that the result is not empty.
// B) Fetch attempts from a user who has no submissions and check that the result is empty.
// C) Make a submission to a user with no entries, then fetch that attempt and makes sure the result has one entry.
// D) Remove that submission and then fetch again and maker sure the result is empty.
func (this *DBTests) DBTestFetchAttempts(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	assignment := MustGetTestAssignment()

	// Case A
	studentAttempts, err := GetSubmissionAttempts(assignment, "course-student@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get student attempts when should be a success: '%v'.", err)
	}

	if len(studentAttempts) == 0 {
		test.Fatalf("Got an empty result when shouldn't be.")
	}

	// Case B
	graderAttempts, err := GetSubmissionAttempts(assignment, "course-grader@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get grader attempts when should be a success (with empty result): '%v'.", err)
	}

	if len(graderAttempts) != 0 {
		test.Fatalf("Unexpected result length. Expected: '%d', Actual: '%d'.", 0, len(graderAttempts))
	}

	// Case C
	graderSubmission := studentAttempts[0]
	graderSubmission.Info.User = "course-grader@test.edulinq.org"

	err = SaveSubmission(assignment, graderSubmission)
	if err != nil {
		test.Fatalf("Failed to save grader submission: '%v'.", err)
	}

	graderAttempts, err = GetSubmissionAttempts(assignment, "course-grader@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get grader attempts when there should be one: '%v'.", err)
	}

	if len(graderAttempts) != 1 {
		test.Fatalf("Fetch returned unexpected number of attempts. Expected: '%d', Actual: '%d'.", 1, len(graderAttempts))
	}

	if !reflect.DeepEqual(graderAttempts[0], graderSubmission) {
		test.Errorf("Unexpected attempt returned. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(graderAttempts[0]), util.MustToJSONIndent(graderSubmission))
	}

	// Case D
	isRemoved, err := RemoveSubmission(assignment, "course-grader@test.edulinq.org", "")
	if err != nil {
		test.Fatalf("Failed to remove grader submission: '%v'.", err)
	}

	if !isRemoved {
		test.Fatalf("Returned false from RemoveSubmission() when should be true.")
	}

	graderAttempts, err = GetSubmissionAttempts(assignment, "course-grader@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get grader attempts when should be a success (with empty result): '%v'.", err)
	}

	if len(graderAttempts) != 0 {
		test.Fatalf("Unexpected result length. Expected: '%d', Actual: '%d'.", 0, len(graderAttempts))
	}
}

// Ensure that stdout and stderr are fetched cleanly.
func (this *DBTests) DBTestGetSubmissionContentsStdoutStderr(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	assignment := MustGetTestAssignment()

	result, err := GetSubmissionContents(assignment, "course-student@test.edulinq.org", "1697406256")
	if err != nil {
		test.Fatalf("Failed to get attempt: '%v'.", err)
	}

	expectedStdout := strings.TrimSpace(baseExpectedStdout)
	expectedStderr := strings.TrimSpace(baseExpectedStderr)

	actualStdout := strings.TrimSpace(result.Stdout)
	actualStderr := strings.TrimSpace(result.Stderr)

	if expectedStdout != actualStdout {
		test.Fatalf("Stdout does not match. Expected: '%s', Actual: '%s'.", expectedStdout, actualStdout)
	}

	if expectedStderr != actualStderr {
		test.Fatalf("Stderr does not match. Expected: '%s', Actual: '%s'.", expectedStderr, actualStderr)
	}
}

func (this *DBTests) DBTestGetSubmissionHistoryBase(test *testing.T) {
	assignment := MustGetTestAssignment()

	results, err := GetSubmissionHistory(assignment, "course-student@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get submission history: '%v'.", err)
	}

	expected := []*model.SubmissionHistoryItem{
		&model.SubmissionHistoryItem{
			ID:               "course101::hw0::course-student@test.edulinq.org::1697406256",
			ShortID:          "1697406256",
			CourseID:         "course101",
			AssignmentID:     "hw0",
			User:             "course-student@test.edulinq.org",
			Message:          "",
			MaxPoints:        2,
			Score:            0,
			GradingStartTime: timestamp.FromMSecs(1697406256000),
		},
		&model.SubmissionHistoryItem{
			ID:               "course101::hw0::course-student@test.edulinq.org::1697406265",
			ShortID:          "1697406265",
			CourseID:         "course101",
			AssignmentID:     "hw0",
			User:             "course-student@test.edulinq.org",
			Message:          "",
			MaxPoints:        2,
			Score:            1,
			GradingStartTime: timestamp.FromMSecs(1697406266000),
		},
		&model.SubmissionHistoryItem{
			ID:               "course101::hw0::course-student@test.edulinq.org::1697406272",
			ShortID:          "1697406272",
			CourseID:         "course101",
			AssignmentID:     "hw0",
			User:             "course-student@test.edulinq.org",
			Message:          "",
			MaxPoints:        2,
			Score:            2,
			GradingStartTime: timestamp.FromMSecs(1697406273000),
		},
	}

	if !reflect.DeepEqual(expected, results) {
		test.Fatalf("Unexpected results. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(results))
	}
}

func (this *DBTests) DBTestGetPreviousSubmissionIDBase(test *testing.T) {
	testCases := []struct {
		targetSubmission string
		expected         string
	}{
		{"ZZZ", ""},

		{"1697406256", ""},
		{"1697406265", "course101::hw0::course-student@test.edulinq.org::1697406256"},
		{"1697406272", "course101::hw0::course-student@test.edulinq.org::1697406265"},

		{"course101::hw0::course-student@test.edulinq.org::1697406256", ""},
		{"course101::hw0::course-student@test.edulinq.org::1697406265", "course101::hw0::course-student@test.edulinq.org::1697406256"},
		{"course101::hw0::course-student@test.edulinq.org::1697406272", "course101::hw0::course-student@test.edulinq.org::1697406265"},
	}

	assignment := MustGetTestAssignment()
	email := "course-student@test.edulinq.org"

	for i, testCase := range testCases {
		previousID, err := GetPreviousSubmissionID(assignment, email, testCase.targetSubmission)
		if err != nil {
			test.Errorf("Case %d: Failed to get previous submission ID: '%v'.", i, err)
			continue
		}

		if testCase.expected != previousID {
			test.Errorf("Case %d: Previous ID not as expected. Expected: '%s', Actual: '%s'.", i, testCase.expected, previousID)
			continue
		}
	}
}

const baseExpectedStdout string = `
Autograder transcript for assignment: HW0.
Grading started at 2023-11-11 22:13 and ended at 2023-11-11 22:13.
Q1: 0 / 1
   NotImplemented returned.
Q2: 0 / 1
   NotImplemented returned.
Style: 0 / 0
   Style is clean!

Total: 0 / 2
`

const baseExpectedStderr string = `
    Dummy Stderr
`

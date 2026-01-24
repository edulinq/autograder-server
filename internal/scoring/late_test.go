package scoring

import (
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Ensure that the late days info struct can be serialized
// (since it is serialized in a Must function).
func TestLateDaysInfoStruct(test *testing.T) {
	testCases := []*LateDaysInfo{
		nil,
		&LateDaysInfo{},
		&LateDaysInfo{
			AvailableDays:           1,
			UploadTime:              timestamp.Now(),
			AllocatedDays:           map[string]int{"A": 1, "B": 2},
			AllocationValues:        map[string]float64{"A": 10.0, "B": 20.0},
			DaysLatePerAssignment:   map[string]int{"A": 1, "B": 2},
			AutograderStructVersion: LATE_DAYS_STRUCT_VERSION,
			LMSCommentID:            "foo",
			LMSCommentAuthorID:      "bar",
		},
	}

	for _, testCase := range testCases {
		util.MustToJSON(testCase)
	}
}

// Ensure that the LateDaysInfo struct can be identified as an autograder comment.
func TestLateDaysInfoJSONContainsAutograderKey(test *testing.T) {
	content := util.MustToJSON(LateDaysInfo{})

	if !strings.Contains(content, common.AUTOGRADER_COMMENT_IDENTITY_KEY) {
		test.Fatalf("JSON does not contain autograder substring '%s': '%s'.",
			common.AUTOGRADER_COMMENT_IDENTITY_KEY, content)
	}
}

func TestComputeLateDays(test *testing.T) {
	var dayMSecs int64 = 24 * 60 * 60 * 1000

	testCases := []struct {
		dueDate        timestamp.Timestamp
		submissionTime timestamp.Timestamp
		expected       int
	}{
		{timestamp.Timestamp(0), timestamp.Timestamp(0), 0},
		{timestamp.Timestamp(0), timestamp.Timestamp(-1), 0},

		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp(dayMSecs), 0},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp(dayMSecs - 1), 0},

		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs), 1},
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + 1), 2},
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs - 1), 1},

		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) - 1), 1},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) + 0), 1},
		{timestamp.Timestamp(dayMSecs), timestamp.Timestamp((2 * dayMSecs) + 1), 2},
	}

	for i, testCase := range testCases {
		actual := computeLateDays(testCase.dueDate, testCase.submissionTime, 0)
		if testCase.expected != actual {
			test.Errorf("Case %d: Bad late days. Expected: %d, Actual: %d.", i, testCase.expected, actual)
		}
	}
}

func TestComputeLateDaysWithGraceTime(test *testing.T) {
	var dayMSecs int64 = 24 * 60 * 60 * 1000
	var minuteMSecs int64 = 60 * 1000

	testCases := []struct {
		dueDate        timestamp.Timestamp
		submissionTime timestamp.Timestamp
		graceMinutes   int
		expected       int
	}{
		// No late time, with grace time should still be 0.
		{timestamp.Timestamp(0), timestamp.Timestamp(0), 10, 0},

		// Submission within grace period should be 0.
		{timestamp.Timestamp(0), timestamp.Timestamp(5 * minuteMSecs), 10, 0},
		{timestamp.Timestamp(0), timestamp.Timestamp(10 * minuteMSecs), 10, 0},

		// Submission just after grace period should be 1 day late.
		{timestamp.Timestamp(0), timestamp.Timestamp((10 * minuteMSecs) + 1), 10, 1},

		// Submission 1 day + grace time should be 1 day late.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + (10 * minuteMSecs)), 10, 1},

		// Submission 1 day + grace time + 1ms should be 2 days late.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs + (10 * minuteMSecs) + 1), 10, 2},

		// Zero grace time should behave like original function.
		{timestamp.Timestamp(0), timestamp.Timestamp(dayMSecs), 0, 1},
		{timestamp.Timestamp(0), timestamp.Timestamp(1), 0, 1},
	}

	for i, testCase := range testCases {
		actual := computeLateDays(testCase.dueDate, testCase.submissionTime, testCase.graceMinutes)
		if testCase.expected != actual {
			test.Errorf("Case %d: Bad late days with grace time. Expected: %d, Actual: %d.", i, testCase.expected, actual)
		}
	}
}

func TestComputeBenevolentAllocation(test *testing.T) {
	testCases := []struct {
		description          string
		lateDays             *LateDaysInfo
		currentAssignmentID  string
		currentDaysLate      int
		currentPenalty       float64
		maxLateDaysPerAssign int
		totalAvailable       int
		expectedAllocation   int
	}{
		{
			description: "Single assignment, enough late days",
			lateDays: &LateDaysInfo{
				AllocatedDays:         make(map[string]int),
				AllocationValues:      make(map[string]float64),
				DaysLatePerAssignment: make(map[string]int),
			},
			currentAssignmentID:  "A",
			currentDaysLate:      2,
			currentPenalty:       10.0,
			maxLateDaysPerAssign: 3,
			totalAvailable:       5,
			expectedAllocation:   2, // Uses 2 days (limited by days late)
		},
		{
			description: "Single assignment, limited by max per assignment",
			lateDays: &LateDaysInfo{
				AllocatedDays:         make(map[string]int),
				AllocationValues:      make(map[string]float64),
				DaysLatePerAssignment: make(map[string]int),
			},
			currentAssignmentID:  "A",
			currentDaysLate:      5,
			currentPenalty:       10.0,
			maxLateDaysPerAssign: 2,
			totalAvailable:       5,
			expectedAllocation:   2, // Limited by max per assignment
		},
		{
			description: "Two assignments, current has higher value",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 5.0}, // Lower value
				DaysLatePerAssignment: map[string]int{"B": 2},
			},
			currentAssignmentID:  "A",
			currentDaysLate:      2,
			currentPenalty:       10.0, // Higher value - should get priority
			maxLateDaysPerAssign: 2,
			totalAvailable:       3, // 3 available (not counting B's allocation)
			expectedAllocation:   2, // A gets 2, B gets reclaimed 2 + 3 - 2 = 3 remaining
		},
		{
			description: "Two assignments, previous has higher value",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 20.0}, // Higher value
				DaysLatePerAssignment: map[string]int{"B": 2},
			},
			currentAssignmentID:  "A",
			currentDaysLate:      2,
			currentPenalty:       5.0, // Lower value
			maxLateDaysPerAssign: 2,
			totalAvailable:       1, // Only 1 available (not counting B's allocation)
			expectedAllocation:   1, // B keeps 2, A gets remaining 1
		},
		{
			description: "Not enough late days, current has lower value",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 15.0}, // Higher value - gets priority
				DaysLatePerAssignment: map[string]int{"B": 2},
			},
			currentAssignmentID:  "A",
			currentDaysLate:      2,
			currentPenalty:       10.0, // Lower value
			maxLateDaysPerAssign: 2,
			totalAvailable:       0, // 0 available (not counting B's allocation)
			expectedAllocation:   0, // Total pool: 2 (from B). B keeps 2 (higher value), A gets 0
		},
		{
			description: "Equal penalty values, current assignment gets priority",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 10.0}, // Same value as current
				DaysLatePerAssignment: map[string]int{"B": 2},
			},
			currentAssignmentID:  "A",
			currentDaysLate:      2,
			currentPenalty:       10.0, // Same value as B
			maxLateDaysPerAssign: 2,
			totalAvailable:       0, // 0 available (not counting B's allocation)
			expectedAllocation:   2, // Total pool: 2. Equal values, A added first so gets priority
		},
	}

	for i, testCase := range testCases {
		actual := computeBenevolentAllocation(
			testCase.lateDays,
			testCase.currentAssignmentID,
			testCase.currentDaysLate,
			testCase.currentPenalty,
			testCase.maxLateDaysPerAssign,
			testCase.totalAvailable,
		)
		if testCase.expectedAllocation != actual {
			test.Errorf("Case %d (%s): Bad benevolent allocation. Expected: %d, Actual: %d.",
				i, testCase.description, testCase.expectedAllocation, actual)
		}
	}
}

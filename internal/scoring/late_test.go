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
			SubmissionTimes:         map[string]timestamp.Timestamp{"A": timestamp.Timestamp(1000), "B": timestamp.Timestamp(2000)},
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

func TestComputeLateDayAllocation(test *testing.T) {
	// Submission times for testing (in milliseconds).
	timeA := timestamp.Timestamp(1000) // Earlier
	timeB := timestamp.Timestamp(2000) // Later

	testCases := []struct {
		description           string
		lateDays              *LateDaysInfo
		currentAssignmentID   string
		currentDaysLate       int
		currentPenalty        float64
		currentSubmissionTime timestamp.Timestamp
		maxLateDaysPerAssign  int
		totalAvailable        int
		expectedStandard      int // Expected result in standard mode
		expectedOptimal       int // Expected result in optimal mode
	}{
		{
			description: "Single assignment, enough late days",
			lateDays: &LateDaysInfo{
				AllocatedDays:         make(map[string]int),
				AllocationValues:      make(map[string]float64),
				DaysLatePerAssignment: make(map[string]int),
				SubmissionTimes:       make(map[string]timestamp.Timestamp),
			},
			currentAssignmentID:   "A",
			currentDaysLate:       2,
			currentPenalty:        10.0,
			currentSubmissionTime: timeA,
			maxLateDaysPerAssign:  3,
			totalAvailable:        5,
			expectedStandard:      2, // Uses 2 days (limited by days late)
			expectedOptimal:       2, // Same result for single assignment
		},
		{
			description: "Single assignment, limited by max per assignment",
			lateDays: &LateDaysInfo{
				AllocatedDays:         make(map[string]int),
				AllocationValues:      make(map[string]float64),
				DaysLatePerAssignment: make(map[string]int),
				SubmissionTimes:       make(map[string]timestamp.Timestamp),
			},
			currentAssignmentID:   "A",
			currentDaysLate:       5,
			currentPenalty:        10.0,
			currentSubmissionTime: timeA,
			maxLateDaysPerAssign:  2,
			totalAvailable:        5,
			expectedStandard:      2, // Limited by max per assignment
			expectedOptimal:       2, // Same result for single assignment
		},
		{
			description: "Two assignments, current submitted later with higher value",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 5.0}, // Lower value
				DaysLatePerAssignment: map[string]int{"B": 2},
				SubmissionTimes:       map[string]timestamp.Timestamp{"B": timeA}, // B submitted earlier
			},
			currentAssignmentID:   "A",
			currentDaysLate:       2,
			currentPenalty:        10.0,  // Higher value
			currentSubmissionTime: timeB, // A submitted later
			maxLateDaysPerAssign:  2,
			totalAvailable:        1, // 1 available (not counting B's allocation)
			expectedStandard:      1, // Standard: B gets 2 (earlier), A gets remaining 1
			expectedOptimal:       2, // Optimal: A gets 2 (higher value), B gets 1
		},
		{
			description: "Two assignments, current submitted earlier with lower value",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 20.0}, // Higher value
				DaysLatePerAssignment: map[string]int{"B": 2},
				SubmissionTimes:       map[string]timestamp.Timestamp{"B": timeB}, // B submitted later
			},
			currentAssignmentID:   "A",
			currentDaysLate:       2,
			currentPenalty:        5.0,   // Lower value
			currentSubmissionTime: timeA, // A submitted earlier
			maxLateDaysPerAssign:  2,
			totalAvailable:        1, // 1 available (not counting B's allocation)
			expectedStandard:      2, // Standard: A gets 2 (earlier), B gets 1
			expectedOptimal:       1, // Optimal: B gets 2 (higher value), A gets 1
		},
		{
			description: "Not enough late days, current has lower value and submitted later",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 15.0}, // Higher value
				DaysLatePerAssignment: map[string]int{"B": 2},
				SubmissionTimes:       map[string]timestamp.Timestamp{"B": timeA}, // B submitted earlier
			},
			currentAssignmentID:   "A",
			currentDaysLate:       2,
			currentPenalty:        10.0,  // Lower value
			currentSubmissionTime: timeB, // A submitted later
			maxLateDaysPerAssign:  2,
			totalAvailable:        0, // 0 available (not counting B's allocation)
			expectedStandard:      0, // Standard: B gets 2 (earlier), A gets 0
			expectedOptimal:       0, // Optimal: B gets 2 (higher value), A gets 0
		},
		{
			description: "Equal penalty values, current submitted earlier",
			lateDays: &LateDaysInfo{
				AllocatedDays:         map[string]int{"B": 2},
				AllocationValues:      map[string]float64{"B": 10.0}, // Same value
				DaysLatePerAssignment: map[string]int{"B": 2},
				SubmissionTimes:       map[string]timestamp.Timestamp{"B": timeB}, // B submitted later
			},
			currentAssignmentID:   "A",
			currentDaysLate:       2,
			currentPenalty:        10.0,  // Same value
			currentSubmissionTime: timeA, // A submitted earlier
			maxLateDaysPerAssign:  2,
			totalAvailable:        0, // 0 available (not counting B's allocation)
			expectedStandard:      2, // Standard: A gets 2 (earlier), B gets 0
			expectedOptimal:       2, // Optimal: A gets 2 (added first), B gets 0
		},
	}

	for i, testCase := range testCases {
		// Test standard mode.
		// Create a fresh copy of lateDays for standard mode test.
		lateDaysStandard := copyLateDaysInfo(testCase.lateDays)
		actualStandard := computeLateDayAllocation(
			lateDaysStandard,
			testCase.currentAssignmentID,
			testCase.currentDaysLate,
			testCase.currentPenalty,
			testCase.currentSubmissionTime,
			testCase.maxLateDaysPerAssign,
			testCase.totalAvailable,
			false, // standard mode
		)
		if testCase.expectedStandard != actualStandard {
			test.Errorf("Case %d (%s) [standard]: Bad allocation. Expected: %d, Actual: %d.",
				i, testCase.description, testCase.expectedStandard, actualStandard)
		}

		// Test optimal mode.
		// Create a fresh copy of lateDays for optimal mode test.
		lateDaysOptimal := copyLateDaysInfo(testCase.lateDays)
		actualOptimal := computeLateDayAllocation(
			lateDaysOptimal,
			testCase.currentAssignmentID,
			testCase.currentDaysLate,
			testCase.currentPenalty,
			testCase.currentSubmissionTime,
			testCase.maxLateDaysPerAssign,
			testCase.totalAvailable,
			true, // optimal mode
		)
		if testCase.expectedOptimal != actualOptimal {
			test.Errorf("Case %d (%s) [optimal]: Bad allocation. Expected: %d, Actual: %d.",
				i, testCase.description, testCase.expectedOptimal, actualOptimal)
		}
	}
}

// copyLateDaysInfo creates a deep copy of LateDaysInfo for testing.
func copyLateDaysInfo(original *LateDaysInfo) *LateDaysInfo {
	copy := &LateDaysInfo{
		AvailableDays:           original.AvailableDays,
		UploadTime:              original.UploadTime,
		AllocatedDays:           make(map[string]int),
		AllocationValues:        make(map[string]float64),
		DaysLatePerAssignment:   make(map[string]int),
		SubmissionTimes:         make(map[string]timestamp.Timestamp),
		AutograderStructVersion: original.AutograderStructVersion,
		LMSCommentID:            original.LMSCommentID,
		LMSCommentAuthorID:      original.LMSCommentAuthorID,
	}

	for k, v := range original.AllocatedDays {
		copy.AllocatedDays[k] = v
	}

	for k, v := range original.AllocationValues {
		copy.AllocationValues[k] = v
	}

	for k, v := range original.DaysLatePerAssignment {
		copy.DaysLatePerAssignment[k] = v
	}

	for k, v := range original.SubmissionTimes {
		copy.SubmissionTimes[k] = v
	}

	return copy
}

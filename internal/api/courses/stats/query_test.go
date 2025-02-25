package stats

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestQuery(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email            string
		permErrorLocator string
		query            stats.CourseMetricQuery
		expectedValues   []int
	}{
		// Base
		{"server-admin", "", stats.CourseMetricQuery{}, []int{100, 200, 300}},
		{"server-admin", "", stats.CourseMetricQuery{BaseQuery: stats.BaseQuery{Sort: 1}}, []int{300, 200, 100}},
		{"server-admin", "", stats.CourseMetricQuery{BaseQuery: stats.BaseQuery{After: timestamp.FromMSecs(150)}}, []int{200, 300}},

		// Course Specific
		{"server-admin", "", stats.CourseMetricQuery{CourseAssignmentEmailQuery: stats.CourseAssignmentEmailQuery{AssignmentID: "A2"}}, []int{200}},
		{"server-admin", "", stats.CourseMetricQuery{CourseAssignmentEmailQuery: stats.CourseAssignmentEmailQuery{AssignmentID: "ZZZ"}}, nil},
		{"server-admin", "", stats.CourseMetricQuery{CourseAssignmentEmailQuery: stats.CourseAssignmentEmailQuery{UserEmail: "U1"}}, []int{100, 200}},
		{"server-admin", "", stats.CourseMetricQuery{Type: stats.CourseMetricTypeGradingTime}, []int{100, 300}},

		// Error
		{"server-user", "-040", stats.CourseMetricQuery{}, nil},
		{"course-student", "-020", stats.CourseMetricQuery{}, nil},
	}

	for _, record := range testRecords {
		err := db.StoreCourseMetric(record)
		if err != nil {
			test.Fatalf("Failed to store test record: '%v'.", err)
		}
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, `courses/stats/query`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.permErrorLocator != "" {
				if testCase.permErrorLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator on perm error. Expected: '%s', Actual: '%s'.", i, testCase.permErrorLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		var responseContent QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if len(testCase.expectedValues) != len(responseContent.Records) {
			test.Errorf("Case %d: Unexpected number of records. Expected: %d, Actual: %d.", i, len(testCase.expectedValues), len(responseContent.Records))
			continue
		}

		match := true
		for i, _ := range responseContent.Records {
			expectedTimestamp := timestamp.FromMSecs(int64(testCase.expectedValues[i]))
			match = (match && (expectedTimestamp == responseContent.Records[i].Timestamp))
		}

		if !match {
			test.Errorf("Case %d: Unexpected record timestamps. Expected: %s, Actual: %s.", i, util.MustToJSONIndent(testCase.expectedValues), util.MustToJSONIndent(responseContent.Records))
			continue
		}
	}
}

var testRecords []*stats.CourseMetric = []*stats.CourseMetric{
	&stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(100),
		},
		CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
			CourseID:     db.TEST_COURSE_ID,
			AssignmentID: "A1",
			UserEmail:    "U1",
		},
		Type:  stats.CourseMetricTypeGradingTime,
		Value: 100,
	},
	&stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(200),
		},
		CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
			CourseID:     db.TEST_COURSE_ID,
			AssignmentID: "A2",
			UserEmail:    "U1",
		},
		Type:  stats.CourseMetricTypeUnknown,
		Value: 200,
	},
	&stats.CourseMetric{
		BaseMetric: stats.BaseMetric{
			Timestamp: timestamp.FromMSecs(300),
		},
		CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
			CourseID:     db.TEST_COURSE_ID,
			AssignmentID: "A3",
			UserEmail:    "U2",
		},
		Type:  stats.CourseMetricTypeGradingTime,
		Value: 300,
	},
}

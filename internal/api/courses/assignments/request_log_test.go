package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

// Test API request logs are properly stored for APIRequestAssignmentContext requests.
func TestRequestLog(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	users := db.MustGetServerUsers()

	testCases := []struct {
		email            string
		endpoint         string
		permErrorLocator string
		query            stats.BaseQuery
		expectedMetric   *stats.APIRequestMetric
	}{
		// Valid permissions.
		{
			email:            "server-admin@test.edulinq.org",
			endpoint:         "courses/assignments/get",
			query:            stats.BaseQuery{},
			permErrorLocator: "",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint:     "/api/v03/courses/assignments/get",
				UserEmail:    "server-admin@test.edulinq.org",
				CourseID:     "course101",
				AssignmentID: "hw0",
			},
		},
		{
			email:            "course-admin@test.edulinq.org",
			endpoint:         "courses/assignments/get",
			query:            stats.BaseQuery{},
			permErrorLocator: "",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint:     "/api/v03/courses/assignments/get",
				UserEmail:    "course-admin@test.edulinq.org",
				CourseID:     "course101",
				AssignmentID: "hw0",
			},
		},

		// Invalid permissions.
		{
			email:            "server-user@test.edulinq.org",
			endpoint:         "courses/assignments/get",
			query:            stats.BaseQuery{},
			permErrorLocator: "-040",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/assignments/get",
				Locator:  "-040",
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(*users[testCase.email].Name),
		}

		response := core.SendTestAPIRequestFull(test, testCase.endpoint, fields, nil, testCase.email)
		if !response.Success {
			if testCase.permErrorLocator != "" {
				if testCase.permErrorLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator on perm error. Expected: '%s', Actual: '%s'.", i, testCase.permErrorLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}
		}

		metric, err := db.GetAPIRequestMetrics(testCase.query)
		if err != nil {
			test.Errorf("Case %d: Unable to get API request metrics: '%v'.", i, err)
			continue
		}

		if len(metric) == 0 {
			test.Errorf("Case %d: No request metrics collected.", i)
			continue
		}

		if metric[0].Timestamp == 0 {
			test.Errorf("Case %d: Timestamp field was not properly populated: '%v'.", i, metric)
			continue
		}

		if metric[0].Sender == "" {
			test.Errorf("Case %d: Sender field was not properly populated: '%v'.", i, metric)
			continue
		}

		if metric[0].Duration == 0 {
			test.Errorf("Case %d: Duration field was not properly populated: '%v'.", i, metric)
		}

		metric[0].Timestamp = 0
		metric[0].Sender = ""
		metric[0].Duration = 0

		if !reflect.DeepEqual(metric[0], testCase.expectedMetric) {
			test.Errorf("Case %d: Stored metric is not as expected. Expected: '%v', Actual: %v", i, util.MustToJSONIndent(testCase.expectedMetric), util.MustToJSONIndent(metric[0]))
			continue
		}
	}
}

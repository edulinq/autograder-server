package users

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

// Test API request logs are properly stored for APIRequestCourseUserContext requests.
func TestRequestLog(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testCases := []struct {
		email            string
		endpoint         string
		permErrorLocator string
		query            stats.BaseQuery
		expectedMetric   *stats.APIRequestMetric
	}{
		// Valid permissions.
		{
			email:            "server-admin",
			endpoint:         "courses/users/list",
			query:            stats.BaseQuery{},
			permErrorLocator: "",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint:  "/api/v03/courses/users/list",
				UserEmail: "server-admin@test.edulinq.org",
				CourseID:  "course101",
			},
		},
		{
			email:            "course-admin",
			endpoint:         "courses/users/list",
			query:            stats.BaseQuery{},
			permErrorLocator: "",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint:  "/api/v03/courses/users/list",
				UserEmail: "course-admin@test.edulinq.org",
				CourseID:  "course101",
			},
		},

		// Invalid permissions.
		{
			email:            "course-student",
			endpoint:         "courses/users/list",
			query:            stats.BaseQuery{},
			permErrorLocator: "-020",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint:  "/api/v03/courses/users/list",
				UserEmail: "course-student@test.edulinq.org",
				CourseID:  "course101",
				Locator:   "-020",
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		response := core.SendTestAPIRequestFull(test, testCase.endpoint, nil, nil, testCase.email)
		if !response.Success {
			if testCase.permErrorLocator != "" {
				if testCase.permErrorLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator on perm error. Expected: '%s', Actual: '%s'.", i, testCase.permErrorLocator, response.Locator)
					continue
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
				continue
			}
		}

		metrics, err := db.GetAPIRequestMetrics(testCase.query)
		if err != nil {
			test.Errorf("Case %d: Unable to get API request metrics: '%v'.", i, err)
			continue
		}

		if len(metrics) == 0 {
			test.Errorf("Case %d: No request metrics collected.", i)
			continue
		}

		// Take first metric since we only make one API request per test case.
		metric := metrics[0]

		if metric.Timestamp == 0 {
			test.Errorf("Case %d: Timestamp field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		if metric.Sender == "" {
			test.Errorf("Case %d: Sender field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		if metric.Duration == 0 {
			test.Errorf("Case %d: Duration field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		// Zero out non-deterministic fields.
		metric.Timestamp = 0
		metric.Sender = ""
		metric.Duration = 0

		if !reflect.DeepEqual(metric, testCase.expectedMetric) {
			test.Errorf("Case %d: Stored metric is not as expected. Expected: '%v', Actual: %v", i, util.MustToJSONIndent(testCase.expectedMetric), util.MustToJSONIndent(metric))
			continue
		}
	}
}

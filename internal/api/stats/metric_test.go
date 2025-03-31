package stats

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

// Test API Request stats get properly stored.
// Normal API tests typically only use the routes defined in their packages,
// but here, multiple routes (defined in ./main_test.go) from different packages are tested.
// This ensures that different types of API requests get their metrics stored properly.
func TestAPIRequestMetrics(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		endpoint        string
		expectedLocator string
		fields          map[string]any
		expectedMetric  *stats.Metric
	}{
		// Valid permissions, APIRequestUserContext request.
		{
			email:    "server-admin",
			endpoint: "users/list",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:  "/api/v03/users/list",
					stats.UserEmailKey: "server-admin@test.edulinq.org",
				},
			},
		},

		// Valid permissions, APIRequestCourseUserContext request.
		{
			email:    "server-admin",
			endpoint: "courses/users/list",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:  "/api/v03/courses/users/list",
					stats.UserEmailKey: "server-admin@test.edulinq.org",
					stats.CourseIDKey:  "course101",
				},
			},
		},

		// Valid permissions, APIRequestAssignmentContext request.
		{
			email:    "server-admin",
			endpoint: "courses/assignments/get",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:     "/api/v03/courses/assignments/get",
					stats.UserEmailKey:    "server-admin@test.edulinq.org",
					stats.CourseIDKey:     "course101",
					stats.AssignmentIDKey: "hw0",
				},
			},
		},

		// Invalid permissions, APIRequestUserContext request.
		{
			email:           "course-student",
			endpoint:        "users/list",
			expectedLocator: "-041",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:  "/api/v03/users/list",
					stats.UserEmailKey: "course-student@test.edulinq.org",
					stats.LocatorKey:   "-041",
				},
			},
		},

		// Invalid permissions, APIRequestCourseUserContext request.
		{
			email:           "course-student",
			endpoint:        "courses/users/list",
			expectedLocator: "-020",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:  "/api/v03/courses/users/list",
					stats.UserEmailKey: "course-student@test.edulinq.org",
					stats.CourseIDKey:  "course101",
					stats.LocatorKey:   "-020",
				},
			},
		},

		// Invalid permissions, APIRequestAssignmentContext request.
		{
			email:           "server-user",
			endpoint:        "courses/assignments/get",
			expectedLocator: "-040",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey: "/api/v03/courses/assignments/get",
					stats.LocatorKey:  "-040",
				},
			},
		},

		// Valid permissions, invalid field request.
		{
			email:           "server-admin",
			endpoint:        "courses/assignments/get",
			fields:          map[string]any{"assignment-id": "zzz"},
			expectedLocator: "-022",
			expectedMetric: &stats.Metric{
				Type: stats.APIRequestStatsType,
				Attributes: map[stats.MetricAttribute]any{
					stats.EndpointKey:     "/api/v03/courses/assignments/get",
					stats.CourseIDKey:     "course101",
					stats.AssignmentIDKey: "zzz",
					stats.LocatorKey:      "-022",
				},
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		response := core.SendTestAPIRequestFull(test, testCase.endpoint, testCase.fields, nil, testCase.email)
		if !response.Success {
			if testCase.expectedLocator != "" {
				if testCase.expectedLocator != response.Locator {
					test.Errorf("Case %d: Incorrect locator. Expected: '%s', Actual: '%s'.", i, testCase.expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.expectedLocator != "" {
			test.Errorf("Case %d: Unexpected success when locator '%s' was expected.", i, testCase.expectedLocator)
			continue
		}

		query := stats.Query{
			Type: stats.APIRequestStatsType,
		}

		metrics, err := db.GetMetrics(query)
		if err != nil {
			test.Errorf("Case %d: Unable to get API request metrics: '%v'.", i, err)
			continue
		}

		if len(metrics) != 1 {
			test.Errorf("Case %d: Got an unexpected number of metrics. Expected: 1, Actual: %d.", i, len(metrics))
			continue
		}

		// Take the first metric since we make one API request per test case.
		metric := metrics[0]

		if metric.Timestamp == 0 {
			test.Errorf("Case %d: Timestamp field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		if metric.Attributes[stats.SenderKey] == "" {
			test.Errorf("Case %d: Sender field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		if metric.Value == 0 {
			test.Errorf("Case %d: Value field was not properly populated: '%v'.", i, util.MustToJSONIndent(metric))
			continue
		}

		// Zero out non-deterministic fields.
		metric.Timestamp = 0
		metric.Value = 0
		delete(metric.Attributes, stats.SenderKey)

		if !reflect.DeepEqual(metric, testCase.expectedMetric) {
			test.Errorf("Case %d: Stored metric is not as expected. Expected: '%v', Actual: '%v'.", i, util.MustToJSONIndent(testCase.expectedMetric), util.MustToJSONIndent(metric))
			continue
		}
	}
}

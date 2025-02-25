package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

// Test API request logs are properly stored.
// This test needs to be in the server package to access all endpoints and share the server testing infrastructure.
func TestAPIRequestLog(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email           string
		endpoint        string
		expectedLocator string
		fields          map[string]any
		expectedMetric  *stats.APIRequestMetric
	}{
		// Valid permissions, APIRequestUserContext request.
		{
			email:    "server-admin",
			endpoint: "users/list",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/users/list",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					UserEmail: "server-admin@test.edulinq.org",
				},
			},
		},

		// Valid permissions, APIRequestCourseUserContext request.
		{
			email:    "server-admin",
			endpoint: "courses/users/list",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/users/list",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					UserEmail: "server-admin@test.edulinq.org",
					CourseID:  "course101",
				},
			},
		},

		// Valid permissions, APIRequestAssignmentContext request.
		{
			email:    "server-admin@test.edulinq.org",
			endpoint: "courses/assignments/get",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/assignments/get",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					UserEmail:    "server-admin@test.edulinq.org",
					CourseID:     "course101",
					AssignmentID: "hw0",
				},
			},
		},

		// Invalid permissions, APIRequestUserContext request.
		{
			email:           "course-student",
			endpoint:        "users/list",
			expectedLocator: "-041",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/users/list",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					UserEmail: "course-student@test.edulinq.org",
				},
				Locator: "-041",
			},
		},

		// Invalid permissions, APIRequestCourseUserContext request.
		{
			email:           "course-student",
			endpoint:        "courses/users/list",
			expectedLocator: "-020",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/users/list",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					UserEmail: "course-student@test.edulinq.org",
					CourseID:  "course101",
				},
				Locator: "-020",
			},
		},

		// Invalid permissions, APIRequestAssignmentContext request.
		{
			email:           "server-user",
			endpoint:        "courses/assignments/get",
			expectedLocator: "-040",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/assignments/get",
				Locator:  "-040",
			},
		},

		// Valid permissions, invalid field request.
		{
			email:           "server-admin",
			endpoint:        "courses/assignments/get",
			fields:          map[string]any{"assignment-id": "zzz"},
			expectedLocator: "-022",
			expectedMetric: &stats.APIRequestMetric{
				Endpoint: "/api/v03/courses/assignments/get",
				CourseAssignmentEmailMetric: stats.CourseAssignmentEmailMetric{
					CourseID:     "course101",
					AssignmentID: "zzz",
				},
				Locator: "-022",
			},
		},
	}

	port, err := util.GetUnusedPort()
	if err != nil {
		test.Fatalf("Failed to get an unused port: '%v'.", err)
	}

	oldPort := config.WEB_HTTP_PORT.Get()
	config.WEB_HTTP_PORT.Set(port)
	defer config.WEB_HTTP_PORT.Set(oldPort)

	// Adjust the server address for test client.
	oldURL := core.SetTestServerURL(fmt.Sprintf("http://127.0.0.1:%d", port))
	defer core.SetTestServerURL(oldURL)

	for i, testCase := range testCases {
		db.ResetForTesting()

		runServerTestBaseFull(test, testCase.endpoint, testCase.fields, testCase.email, testCase.expectedLocator, fmt.Sprintf("Case %d: ", i))

		metrics, err := db.GetAPIRequestMetrics(stats.APIRequestMetricQuery{})
		if err != nil {
			test.Errorf("Case %d: Unable to get API request metrics: '%v'.", i, err)
			continue
		}

		if len(metrics) != 1 {
			test.Errorf("Case %d: Got an unexpected number of metrics. Expected: 1, Actual: %d.", i, len(metrics))
			continue
		}

		// Take the first metric since we only make one API request per test case.
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

package submission

import (
    "reflect"
    "testing"

    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/report"
    "github.com/edulinq/autograder/util"
)

func TestCourseReport(test *testing.T) {
    testCases := []struct{
        role model.UserRole
        permError bool
        result *FetchCourseReportResponse
    }{
        // Admin
        {model.RoleAdmin, false, &FetchCourseReportResponse{CourseReport: report.TestCourseReport}},

        // Grader
        {model.RoleGrader, false, &FetchCourseReportResponse{CourseReport: report.TestCourseReport}},

        // Student 
        {model.RoleStudent, true, nil},
    };

    for i, testCase := range testCases {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/course-report`), nil, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-020";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                        i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }
            continue;
        }

        var responseContent *FetchCourseReportResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        var expected_marshaled *FetchCourseReportResponse;
        util.MustJSONFromString(util.MustToJSON(testCase.result), &expected_marshaled);

        if (!reflect.DeepEqual(expected_marshaled, responseContent)) {
            test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
                util.MustToJSONIndent(expected_marshaled), util.MustToJSONIndent(responseContent));
            continue;
        }
    }
}

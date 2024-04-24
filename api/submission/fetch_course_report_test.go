package submission

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/api/core"
	"github.com/edulinq/autograder/common"
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
        {model.RoleAdmin,false,expected},
        // Student 
        {model.RoleStudent,true,&FetchCourseReportResponse{}},

    };
    for i, testCase := range testCases {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/course-report`), nil, nil, testCase.role );
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
        var responseContent *FetchCourseReportResponse
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!reflect.DeepEqual(testCase.result, responseContent)) {
            test.Errorf("Case %d: Unexpected submission result. Expected: '%s', actual: '%s'.", i,
                util.MustToJSONIndent(testCase.result), util.MustToJSONIndent(responseContent));
            continue;
        }
    }


}


var expected = &FetchCourseReportResponse{
    CourseReport: &report.CourseScoringReport{
        CourseName: "Course 101",
        Assignments: []*report.AssignmentScoringReport{
            &report.AssignmentScoringReport{
                AssignmentName: "Homework 0",
                NumberOfSubmissions: 1,
                LatestSubmission: common.MustTimestampFromString("2023-10-15T21:44:33Z"),
                Questions: []*report.ScoringReportQuestionStats{
                    &report.ScoringReportQuestionStats{
                        QuestionName: "Q1",
                        Min: 1,
                        Max: 1,
                        Median: 1,
                        Mean: 1,
                        StdDev: -1,

                        // MinString: "1.00",
                        // MaxString: "1.00",
                        // MedianString: "1.00",
                        // MeanString: "1.00",
                        // StdDevString: "NaN",

                    },
                    &report.ScoringReportQuestionStats{
                        QuestionName: "Q2",
                        Min: 1,
                        Max: 1,
                        Median: 1,
                        Mean: 1,
                        StdDev: -1,

                        // MinString: "1.00",
                        // MaxString: "1.00",
                        // MedianString: "1.00",
                        // MeanString: "1.00",
                        // StdDevString: "NaN",
                    },
                    &report.ScoringReportQuestionStats{
                        QuestionName: "Style",
                        Min: 0,
                        Max: 0,
                        Median: 0,
                        Mean: 0,
                        StdDev: -1,

                        // MinString: "0.00",
                        // MaxString: "0.00",
                        // MedianString: "0.00",
                        // MeanString: "0.00",
                        // StdDevString: "NaN",
                    },
                    &report.ScoringReportQuestionStats{
                        QuestionName: "<Overall>",
                        Min: 1,
                        Max: 1,
                        Median: 1,
                        Mean: 1,
                        StdDev: -1,

                        // MinString: "1.00",
                        // MaxString: "1.00",
                        // MedianString: "1.00",
                        // MeanString: "1.00",
                        // StdDevString: "NaN",
                    },
                },
            },
        },
    },
};
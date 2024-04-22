package submission
import (
    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/model"
    //"github.com/edulinq/autograder/report"
    "testing"
	"fmt"

)

func TestCourseReport(test *testing.T) {

    response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/course-report`), nil, nil, model.RoleAdmin );

    
    fmt.Print(response);

}


// var expected *report.CourseScoringReport = &report.CourseScoringReport{
//     CourseName: "Course 101",
//     Assignments: []*AssignmentScoringReport{
//         &AssignmentScoringReport{
//             AssignmentName: "Homework 0",
//             NumberOfSubmissions: 1,
//             LatestSubmission: common.MustTimestampFromString("2023-10-15T21:44:33Z"),
//             Questions: []*ScoringReportQuestionStats{
//                 &ScoringReportQuestionStats{
//                     QuestionName: "Q1",
//                     Min: 1,
//                     Max: 1,
//                     Median: 1,
//                     Mean: 1,
//                     StdDev: -1,
//                     MinString: "1.00",
//                     MaxString: "1.00",
//                     MedianString: "1.00",
//                     MeanString: "1.00",
//                     StdDevString: "NaN",
//                 },
//                 &ScoringReportQuestionStats{
//                     QuestionName: "Q2",
//                     Min: 1,
//                     Max: 1,
//                     Median: 1,
//                     Mean: 1,
//                     StdDev: -1,
//                     MinString: "1.00",
//                     MaxString: "1.00",
//                     MedianString: "1.00",
//                     MeanString: "1.00",
//                     StdDevString: "NaN",
//                 },
//                 &ScoringReportQuestionStats{
//                     QuestionName: "Style",
//                     Min: 0,
//                     Max: 0,
//                     Median: 0,
//                     Mean: 0,
//                     StdDev: -1,
//                     MinString: "0.00",
//                     MaxString: "0.00",
//                     MedianString: "0.00",
//                     MeanString: "0.00",
//                     StdDevString: "NaN",
//                 },
//                 &ScoringReportQuestionStats{
//                     QuestionName: "<Overall>",
//                     Min: 1,
//                     Max: 1,
//                     Median: 1,
//                     Mean: 1,
//                     StdDev: -1,
//                     MinString: "1.00",
//                     MaxString: "1.00",
//                     MedianString: "1.00",
//                     MeanString: "1.00",
//                     StdDevString: "NaN",
//                 },
//             },
//         },
//     },
// };

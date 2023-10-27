package submission

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/usr"
)

func TestSubmit(test *testing.T) {
    testSubmissions, err := grader.GetTestSubmissions(config.COURSES_ROOT.GetString());
    if (err != nil) {
        test.Fatalf("Failed to get test submissions: '%v'.", err);
    }

    for i, testSubmission := range testSubmissions {
        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/submit`), nil, testSubmission.Files, usr.Student);
        if (!response.Success) {
            test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            continue;
        }

        var responseContent SubmitResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!responseContent.GradingSucess) {
            test.Errorf("Case %d: Response is not a grading success when it should be: '%v'.", i, responseContent);
            continue;
        }

        if (!responseContent.SubmissionResult.Equals(testSubmission.TestSubmission.Result, !testSubmission.TestSubmission.IgnoreMessages)) {
            test.Errorf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
                    responseContent.SubmissionResult, &testSubmission.TestSubmission.Result);
            continue;
        }
    }
}

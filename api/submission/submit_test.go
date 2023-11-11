package submission

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/util"
    "github.com/eriq-augustine/autograder/usr"
)

func TestSubmit(test *testing.T) {
    testSubmissions, err := grader.GetTestSubmissions(config.COURSES_ROOT.Get());
    if (err != nil) {
        test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.COURSES_ROOT.Get(), err);
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

        if (!responseContent.GradingInfo.Equals(*testSubmission.TestSubmission.GradingInfo, !testSubmission.TestSubmission.IgnoreMessages)) {
            test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
                    i, responseContent.GradingInfo, testSubmission.TestSubmission.GradingInfo);
            continue;
        }

        // Fetch the most recent submission from the DB and ensure it matches.
        submission, err := db.GetSubmissionResult(testSubmission.Assignment, "student@test.com", "");
        if (err != nil) {
            test.Errorf("Case %d: Failed to get submission: '%v'.", i, err);
            continue;
        }

        if (!responseContent.GradingInfo.Equals(*submission, !testSubmission.TestSubmission.IgnoreMessages)) {
            test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match database value:\n---\n%v\n---\n.",
                    i, responseContent.GradingInfo, submission);
            continue;
        }
    }
}

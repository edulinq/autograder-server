package images

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/util"
)

func TestInfoBase(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	assignment := db.MustGetTestAssignment()
	expectedImageInfo := assignment.GetImageInfo()

	// Ensure that the image is built.
	err := docker.BuildImageFromSourceQuick(assignment)
	if err != nil {
		test.Fatalf("Could not build image for test: '%v'.", err)
	}

	testCases := []struct {
		User    string
		Locator string
	}{
		// Base
		{
			User: "course-grader",
		},
		{
			User: "server-admin",
		},

		// Invalid Permissions
		{
			User:    "course-student",
			Locator: "-020",
		},
		{
			User:    "server-user",
			Locator: "-040",
		},
	}

	for i, testCase := range testCases {
		response := core.SendTestAPIRequestFull(test, "courses/assignments/images/info", nil, nil, testCase.User)
		if !response.Success {
			if testCase.Locator != response.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected: '%s', Actual: '%s'.",
					i, testCase.Locator, response.Locator)
			}

			continue
		}

		if testCase.Locator != "" {
			test.Errorf("Case %d: Did not get an expected error. Expected: '%s'", i, testCase.Locator)
			continue
		}

		var responseContent InfoResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if assignment.GetImageName() != responseContent.ImageInfo.Name {
			test.Errorf("Case %d: Image name not as expected. Expected: '%s', Actual: '%s'.",
				i, assignment.GetImageName(), responseContent.ImageInfo.Name)
			continue
		}

		// Compare JSON instead of raw structs because of hidden fields.
		expectedImageInfoJSON := util.MustToJSONIndent(expectedImageInfo)
		actualImageInfoJSON := util.MustToJSONIndent(responseContent.ImageInfo.SourceInfo)

		if expectedImageInfoJSON != actualImageInfoJSON {
			test.Errorf("Case %d: Image source info not as expected. Expected: '%s', Actual: '%s'.",
				i, expectedImageInfoJSON, actualImageInfoJSON)
			continue
		}

		minSizeBytes := int64(50 * 1024 * 1024)
		if responseContent.ImageInfo.Size < minSizeBytes {
			test.Errorf("Case %d: Image size less than expected. Expected: '%v', Actual: '%v'.",
				i, minSizeBytes, responseContent.ImageInfo.Size)
			continue
		}
	}
}

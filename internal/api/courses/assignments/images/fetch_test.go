package images

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/util"
)

func TestFetch(test *testing.T) {
	docker.EnsureOrSkipForTest(test)

	testCases := []struct {
		User    string
		Locator string
	}{
		// Base
		{
			User: "course-grader",
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
		response := core.SendTestAPIRequestFull(test, "courses/assignments/images/fetch", nil, nil, testCase.User)
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

		var responseContent FetchResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !responseContent.Gzip {
			test.Errorf("Case %d: Gzip is not true.", i)
		}

		minSizeBytes := int64(50 * 1024 * 1024)
		if responseContent.ImageInfo.Size < minSizeBytes {
			test.Errorf("Case %d: Image size less than expected. Expected: '%v', Actual: '%v'.",
				i, minSizeBytes, responseContent.ImageInfo.Size)
		}

		minGzipSizeBytes := int64(20 * 1024 * 1024)
		if responseContent.ImageInfo.GzipSize < minGzipSizeBytes {
			test.Errorf("Case %d: Image gzip size less than expected. Expected: '%v', Actual: '%v'.",
				i, minGzipSizeBytes, responseContent.ImageInfo.GzipSize)
		}

		if int64(len(responseContent.Bytes)) < minGzipSizeBytes {
			test.Errorf("Case %d: Image byte length less than expected. Expected: '%v', Actual: '%v'.",
				i, minGzipSizeBytes, len(responseContent.Bytes))
		}
	}
}

package server

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/edulinq/autograder/internal/config"
)

func TestMakeHTTPSRedirectBase(test *testing.T) {
	httpsPort := config.WEB_HTTPS_PORT.Get()

	testCases := []struct {
		input    *http.Request
		expected string
	}{
		{
			input: &http.Request{
				URL: &url.URL{
					Scheme: "http",
					Host:   "test.edulinq.org",
					Path:   "/test",
				},
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Host: "test.edulinq.org",
					Path: "/test",
				},
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Path: "/test",
				},
				Host: "test.edulinq.org",
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Host: "test.edulinq.org:1234567890",
					Path: "/test",
				},
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Path: "/test",
				},
				Host: "test.edulinq.org:12345678",
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Path: "/test",
				},
				Host: "[test.edulinq.org]:12345678",
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
		{
			input: &http.Request{
				URL: &url.URL{
					Path: "/test",
					Host: "[test.edulinq.org]:12345678",
				},
			},
			expected: fmt.Sprintf("https://test.edulinq.org:%d/test", httpsPort),
		},
	}

	for i, testCase := range testCases {
		actual, err := makeHTTPSRedirect(testCase.input)
		if err != nil {
			test.Errorf("Case %d: Unexpected error: '%v'.", i, err)
			continue
		}

		if testCase.expected != actual {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', Actual: '%s'.", i, testCase.expected, actual)
			continue
		}
	}
}

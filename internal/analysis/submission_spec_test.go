package analysis

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/db"
)

func TestResolveSubmissionSpecsBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	AddTestSubmissions(test)

	testCases := []struct {
		input                  []string
		expectedIDs            []string
		expectedCourses        []string
		expectedErrorSubstring string
	}{
		{
			[]string{"course101::hw0::course-student@test.edulinq.org::1697406256"},
			[]string{"course101::hw0::course-student@test.edulinq.org::1697406256"},
			[]string{"course101"},
			"",
		},
		{
			[]string{"course101::hw0::course-student@test.edulinq.org"},
			[]string{"course101::hw0::course-student@test.edulinq.org::1697406272"},
			[]string{"course101"},
			"",
		},
		{
			[]string{"course101::hw0"},
			[]string{
				"course101::hw0::course-student-alt@test.edulinq.org::1234567890",
				"course101::hw0::course-student@test.edulinq.org::1697406272",
			},
			[]string{"course101"},
			"",
		},

		{
			[]string{
				"course-languages::bash::course-student-alt@test.edulinq.org::1234567890",
				"course101::hw0::course-student@test.edulinq.org::1697406256",
			},
			[]string{
				"course-languages::bash::course-student-alt@test.edulinq.org::1234567890",
				"course101::hw0::course-student@test.edulinq.org::1697406256",
			},
			[]string{
				"course-languages",
				"course101",
			},
			"",
		},

		{
			[]string{
				// Note the ordering.
				"course101::hw0::course-student@test.edulinq.org::1697406265",
				"course101::hw0::course-student@test.edulinq.org::1697406256",
			},
			[]string{
				"course101::hw0::course-student@test.edulinq.org::1697406256",
				"course101::hw0::course-student@test.edulinq.org::1697406265",
			},
			[]string{"course101"},
			"",
		},

		// Errors

		{
			[]string{"course101::hw0::course-student@test.edulinq.org::ZZZ"},
			nil,
			nil,
			"Submission short ID is invalid",
		},
		{
			[]string{"course101::hw0!!!"},
			nil,
			nil,
			"Assignment ID is invalid",
		},
		{
			[]string{"course101!!!::hw0"},
			nil,
			nil,
			"Course ID is invalid",
		},
		{
			[]string{"course101::ZZZ::course-student@test.edulinq.org::1697406256"},
			nil,
			nil,
			"Unable to find assignment",
		},
		{
			[]string{"ZZZ::hw0::course-student@test.edulinq.org::1697406256"},
			nil,
			nil,
			"Unable to find course",
		},
		{
			[]string{"course101"},
			nil,
			nil,
			"Submission spec has too few components",
		},
		{
			[]string{"A::B::C::D::E"},
			nil,
			nil,
			"Submission spec has too many components",
		},
	}

	for i, testCase := range testCases {
		ids, courses, err := ResolveSubmissionSpecs(testCase.input)
		if err != nil {
			if testCase.expectedErrorSubstring == "" {
				test.Errorf("Case %d: Unexpected error: '%v'.", i, err)
				continue
			}

			if !strings.Contains(err.Error(), testCase.expectedErrorSubstring) {
				test.Errorf("Case %d: Error does not contain expected substring. Substring: '%s', Error: '%s'.",
					i, testCase.expectedErrorSubstring, err.Error())
				continue
			}

			continue
		}

		if testCase.expectedErrorSubstring != "" {
			test.Errorf("Case %d: Did not get an expected error with substring '%s'.", i, testCase.expectedErrorSubstring)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedIDs, ids) {
			test.Errorf("Case %d: Unexpected ids. Expected: '%v', Actual: '%v'.",
				i, testCase.expectedIDs, ids)
			continue
		}

		if !reflect.DeepEqual(testCase.expectedCourses, courses) {
			test.Errorf("Case %d: Unexpected courses. Expected: '%v', Actual: '%v'.",
				i, testCase.expectedCourses, courses)
			continue
		}
	}
}

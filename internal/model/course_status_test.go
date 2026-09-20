package model

import (
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

func TestGetActiveStatus(test *testing.T) {
	testCases := []struct {
		statuses      map[string]*CourseStatus
		expectedOwner string
	}{
		// No statuses, so the getter returns nil.
		{
			map[string]*CourseStatus{},
			"",
		},

		// Single Status
		{
			map[string]*CourseStatus{
				"course-admin@test.edulinq.org": {
					Source:  StatusSourceCourse,
					Owner:   "course-admin@test.edulinq.org",
					SetTime: timestamp.Now(),
				},
			},
			"course-admin@test.edulinq.org",
		},

		// Conflicting Sources
		{
			map[string]*CourseStatus{
				"course-admin@test.edulinq.org": {
					Source:  StatusSourceCourse,
					Owner:   "course-admin@test.edulinq.org",
					SetTime: timestamp.Now(),
				},
				"server-admin@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-admin@test.edulinq.org",
					SetTime: timestamp.Now(),
				},
			},
			"server-admin@test.edulinq.org",
		},
		{
			map[string]*CourseStatus{
				"automated@test.edulinq.org": {
					Source:  StatusSourceAutomated,
					Owner:   "automated@test.edulinq.org",
					SetTime: timestamp.Now(),
				},
				"server-owner@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-owner@test.edulinq.org",
					SetTime: timestamp.Now(),
				},
			},
			"automated@test.edulinq.org",
		},

		// Tie on the source (picks most recent).
		{
			map[string]*CourseStatus{
				"server-owner@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-owner@test.edulinq.org",
					SetTime: timestamp.Zero(),
				},
				"server-admin@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-admin@test.edulinq.org",
					SetTime: timestamp.FromMSecs(100),
				},
			},
			"server-admin@test.edulinq.org",
		},

		// Tie on the source and time (picks owner string that comes last in lexicographical order).
		{
			map[string]*CourseStatus{
				"server-owner@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-owner@test.edulinq.org",
					SetTime: timestamp.FromMSecs(100),
				},
				"server-admin@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "server-admin@test.edulinq.org",
					SetTime: timestamp.FromMSecs(100),
				},
				"z-server-owner@test.edulinq.org": {
					Source:  StatusSourceServer,
					Owner:   "z-server-owner@test.edulinq.org",
					SetTime: timestamp.FromMSecs(100),
				},
			},
			"z-server-owner@test.edulinq.org",
		},
	}

	for i, testCase := range testCases {
		course := &Course{
			Statuses: testCase.statuses,
		}

		result := course.GetActiveStatus()

		if result == nil {
			if testCase.expectedOwner != "" {
				test.Errorf("Case %d: Expected owner '%s', found empty statuses.",
					i, testCase.expectedOwner)
			}
			continue
		}

		if result.Owner != testCase.expectedOwner {
			test.Errorf("Case %d: Expected owner '%s', found '%s'.",
				i, testCase.expectedOwner, result.Owner)
		}
	}
}

package grader

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRegradeBase(test *testing.T) {
	defer db.ResetForTesting()

	studentGradingResults := map[string]*model.GradingResult{
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	testCases := []struct {
		users             []model.CourseUserReference
		waitForCompletion bool
		numLeft           int
		results           map[string]*model.SubmissionHistoryItem
	}{
		// User With Submission, Wait
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": studentGradingResults["1697406272"].Info.ToHistoryItem(),
			},
		},

		// Empty Users, Wait
		{
			[]model.CourseUserReference{},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Submissions, Wait
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": nil,
			},
		},

		// User With Submission, No Wait
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Users, No Wait
		{
			nil,
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{},
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{"-*"},
			false,
			0,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Submission, No Wait
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			false,
			1,
			map[string]*model.SubmissionHistoryItem{},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()
		dummyAssignment := loadDummyAssignment(test)

		options := RegradeOptions{
			GradeOptions: GetDefaultGradeOptions(),
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: testCase.waitForCompletion,
			},
			RawReferences: testCase.users,
			// TODO: Make this a test case field.
			RegradeAfter:          nil,
			RetainOriginalContext: false,
		}

		result, numLeft, err := Regrade(dummyAssignment, options)
		if err != nil {
			test.Errorf("Case %d: Failed to regrade submissions: '%v'.", i, err)
			continue
		}

		if len(result.WorkErrors) != 0 {
			test.Errorf("Case %d: Unexpected work errors during regrade: '%s'.", i, util.MustToJSONIndent(result.WorkErrors))
			continue
		}

		// TODO: Add a check for result.RegradeAfter.

		if testCase.numLeft != numLeft {
			test.Errorf("Case %d: Unexpected number of regrades remaining. Expected: '%d', actual: '%d'.", i, testCase.numLeft, numLeft)
			continue
		}

		failed := CheckAndClearIDs(test, i, testCase.results, result.Results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.results, result.Results) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.results), util.MustToJSONIndent(result.Results))
			continue
		}
	}
}

func loadDummyAssignment(test *testing.T) *model.Assignment {
	dummyCourse := &model.Course{
		ID:          "dummy-course",
		Name:        "Dummy Course",
		Assignments: map[string]*model.Assignment{},
	}

	dummyAssignment := &model.Assignment{
		Course:    dummyCourse,
		ID:        "dummyAssignment",
		Name:      "Dummy Assignment",
		MaxPoints: 1,
	}

	dummyCourse.AddAssignment(dummyAssignment)
	db.SaveCourse(dummyCourse)

	users := map[string]*model.ServerUser{
		"course-admin@test.edulinq.org": &model.ServerUser{
			Email: "course-admin@test.edulinq.org",
			CourseInfo: map[string]*model.UserCourseInfo{
				dummyCourse.GetID(): &model.UserCourseInfo{
					Role: model.CourseRoleAdmin,
				},
			},
		},
		"course-grader@test.edulinq.org": &model.ServerUser{
			Email: "course-grader@test.edulinq.org",
			CourseInfo: map[string]*model.UserCourseInfo{
				dummyCourse.GetID(): &model.UserCourseInfo{
					Role: model.CourseRoleGrader,
				},
			},
		},
		"course-other@test.edulinq.org": &model.ServerUser{
			Email: "course-other@test.edulinq.org",
			CourseInfo: map[string]*model.UserCourseInfo{
				dummyCourse.GetID(): &model.UserCourseInfo{
					Role: model.CourseRoleOther,
				},
			},
		},
		"course-owner@test.edulinq.org": &model.ServerUser{
			Email: "course-owner@test.edulinq.org",
			CourseInfo: map[string]*model.UserCourseInfo{
				dummyCourse.GetID(): &model.UserCourseInfo{
					Role: model.CourseRoleOwner,
				},
			},
		},
		"course-student@test.edulinq.org": &model.ServerUser{
			Email: "course-student@test.edulinq.org",
			CourseInfo: map[string]*model.UserCourseInfo{
				dummyCourse.GetID(): &model.UserCourseInfo{
					Role: model.CourseRoleStudent,
				},
			},
		},
	}

	err := db.UpsertUsers(users)
	if err != nil {
		test.Fatalf("Failed to upsert users into the new course: '%v'.", err)
	}

	return dummyAssignment
}

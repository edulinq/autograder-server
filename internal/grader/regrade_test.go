package grader

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

const DUMMY_COURSE_CONFIG = `{
    "id": "dummy-course",
    "name": "A dummy course to test regrades.",
    "lms": {
        "type": "test"
    }
}`

const DUMMY_ASSIGNMENT_CONFIG = `{
    "id": "dummy-assignment",
    "name": "A dummy assignment to test regrades.",
    "due-date": 0,
    "static-files": [
        "assignment.json",
        "grader.sh"
    ],
    "invocation": ["bash", "./grader.sh"],
    "post-submission-file-ops": [
        ["cp", "input/assignment.sh", "work/assignment.sh"]
    ]
}`

const BASE_GRADER = `#!/bin/bash

readonly THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly DEFAULT_OUTPUT_PATH="${THIS_DIR}/../output/result.json"

function main() {
    if [[ $# -ne 0 ]]; then
        echo "USAGE: $0"
        return 1
    fi

    trap exit SIGINT

    cd "${THIS_DIR}"

    # Allow the grader to run locally by changing the output location
    # if not in the docker image.
    local outputPath="${DEFAULT_OUTPUT_PATH}"
    if [[ ! -d $(dirname "${outputPath}") ]] ; then
        outputPath="$(basename "${outputPath}")"
    fi

    # Run grader.
    grade "${outputPath}"
    if [[ $? -ne 0 ]] ; then
        echo "ERROR: Failed to run grader."
        return 3
    fi

    return 0
}

function grade() {
    # Source the student's assignment file.
    source "${THIS_DIR}/assignment.sh"

    local score=2

    test_eq_one 1 true || { score=$((score-1)); }
    test_eq_one 2 false || { score=$((score-1)); }

    local json_output='{
        "name": "dummy-assignment",
        "questions": [
            {
                "name": "Task 1: eq_one()",
                "max_points": 2,
                "score": '"$score"',
                "message": "'"$message"'"
            }
        ]
    }'

    echo "$json_output" > "${outputPath}"
}`

const FAULTY_GRADER = `
function test_eq_one() {
    local a=$1
    local expected=$2

    local result=$(eq_one $a)
    # BUG in grader!
    [[ $result -ne $expected ]]
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"`

const GOOD_GRADER = `
function test_eq_one() {
    local a=$1
    local expected=$2

    local result=$(eq_one $a)
    # Fixed the bug in the grader.
    [[ $result -eq $expected ]]
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"`

const SUBMISSION = `
function eq_one() {
    local a=$1

    echo $(($a -eq 1))
}`

/*
func TestRegradeBase(test *testing.T) {
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
}
*/

func TestRegradeBase(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		users              []model.CourseUserReference
		initialSubmissions map[string]float64
		waitForCompletion  bool
		numLeft            int
		finalResults       map[string]*model.SubmissionHistoryItem
	}{
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			map[string]float64{
				"course-student@test.edulinq.org": 0,
			},
			true,
			0,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "dummy-course",
					AssignmentID: "assignment-id",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    2,
					Score:        2,
				},
			},
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		tempDir, err := util.MkDirTemp("regrade-grading-dir-")
		if err != nil {
			test.Errorf("Case %d: Failed to create temp dir for grading: '%v'.", i, err)
			continue
		}
		defer util.RemoveDirent(tempDir)

		assignment, err := prepDummyAssignment(tempDir)
		if err != nil {
			test.Errorf("Case %d: Failed to prep dummy assignment: '%v'.", i, err)
			continue
		}

		// Set a relative source dir for the assignment to pass validation.
		assignment.RelSourceDir = "dummy-assignment"
		course := assignment.GetCourse()
		course.Assignments["dummyAssignment"] = assignment
		db.SaveCourse(course)

		gradeOptions := GetDefaultGradeOptions()
		gradeOptions.NoDocker = true
		gradeOptions.CheckRejection = false

		ok := true
		for user, expectedScore := range testCase.initialSubmissions {
			result, reject, softError, err := Grade(context.Background(), assignment, tempDir, user, "", gradeOptions)
			if err != nil {
				test.Errorf("Case %d: Failed to grade initial submission for user '%s': '%v'.", i, user, err)
				ok = false
				break
			}

			if softError != "" {
				test.Errorf("Case %d: Unexpected soft error for initial submission for user '%s': '%s'.", i, user, softError)
				ok = false
				break
			}

			if reject != nil {
				test.Errorf("Case %d: Unexpected rejection during initial submission for user '%s': '%s'.", i, user, reject.String())
				ok = false
				break
			}

			if result.Info.Score != expectedScore {
				test.Errorf("Case %d: Unexpected score on initial submission. Expected: '%f', Actual: '%f'.", i, expectedScore, result.Info.Score)
				ok = false
				break
			}
		}

		if !ok {
			continue
		}

		goodGrader := BASE_GRADER + GOOD_GRADER
		assignmentDir := filepath.Join(tempDir, "dummy-assignment")

		err = util.WriteFile(goodGrader, filepath.Join(assignmentDir, "grader.sh"))
		if err != nil {
			test.Errorf("Case %d: Failed to write good grader: '%v'.", i, err)
			continue
		}

		// TODO: Failing to upsert when the existing course exists.
		err = upsertTestCourseInTmpDir(tempDir)
		if err != nil {
			test.Errorf("Case %d: Failed to upsert test course: '%v'.", i, err)
			continue
		}

		options := RegradeOptions{
			GradeOptions: gradeOptions,
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: testCase.waitForCompletion,
			},
			RawReferences: testCase.users,
			// TODO: Make this a test case field.
			RegradeAfter:          nil,
			RetainOriginalContext: false,
		}

		result, numLeft, err := Regrade(assignment, options)
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

		failed := CheckAndClearIDs(test, i, testCase.finalResults, result.Results)
		if failed {
			continue
		}

		if !reflect.DeepEqual(testCase.finalResults, result.Results) {
			test.Errorf("Case %d: Unexpected regrade result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.finalResults), util.MustToJSONIndent(result.Results))
			continue
		}
	}
}

func prepDummyAssignment(tempDir string) (*model.Assignment, error) {
	err := util.WriteFile(DUMMY_COURSE_CONFIG, filepath.Join(tempDir, "course.json"))
	if err != nil {
		return nil, fmt.Errorf("Failed to write course config: '%v'.", err)
	}

	assignmentDir := filepath.Join(tempDir, "dummy-assignment")
	util.MustMkDir(assignmentDir)

	err = util.WriteFile(DUMMY_ASSIGNMENT_CONFIG, filepath.Join(assignmentDir, "assignment.json"))
	if err != nil {
		return nil, fmt.Errorf("Failed to write assignment config: '%v'.", err)
	}

	faultyGrader := BASE_GRADER + FAULTY_GRADER
	err = util.WriteFile(faultyGrader, filepath.Join(assignmentDir, "grader.sh"))
	if err != nil {
		return nil, fmt.Errorf("Failed to write faulty grader: '%v'.", err)
	}

	err = upsertTestCourseInTmpDir(tempDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to upsert test course: '%v'.", err)
	}

	dummyCourse, err := db.GetCourse("dummy-course")
	if err != nil {
		return nil, fmt.Errorf("Failed to get newly upserted course: '%v'.", err)
	}

	newUsers := map[string]*model.ServerUser{
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

	err = db.UpsertUsers(newUsers)
	if err != nil {
		return nil, fmt.Errorf("Failed to upsert users into the new course: '%v'.", err)
	}

	err = util.WriteFile(SUBMISSION, filepath.Join(tempDir, "assignment.sh"))
	if err != nil {
		return nil, fmt.Errorf("Failed to write submission file: '%v'.", err)
	}

	dummyAssignment := dummyCourse.GetAssignment("dummy-assignment")
	if dummyAssignment == nil {
		return nil, fmt.Errorf("Failed to get dummy assignment from dummy course: '%v'.", err)
	}

	return dummyAssignment, nil
}

func upsertTestCourseInTmpDir(tempDir string) error {
	users, err := db.GetServerUsers()
	if err != nil {
		return fmt.Errorf("Failed to get server users: '%v'.", err)
	}

	upsertOptions := courses.CourseUpsertOptions{
		ContextUser: users["server-admin@test.edulinq.org"],
	}

	_, err = courses.UpsertFromDir(tempDir, upsertOptions)
	if err != nil {
		return fmt.Errorf("Failed to upsert course from dir: '%v'.", err)
	}

	return nil
}

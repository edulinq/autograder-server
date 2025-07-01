package grader

import (
	// TEST
	"fmt"
	"os"

	"context"
	"path/filepath"
	// "reflect"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	// "github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// TODO: Do we need a dummy course config?
// const DUMMY_COURSE_CONFIG = ``

const DUMMY_ASSIGNMENT_CONFIG = `{
    "id": "dummy-assignment",
    "name": "A dummy assignment to test regrades.",
    "due-date": 0,
    "static-files": [
        "grader.sh"
    ],
    "invocation": ["bash", "./grader.sh"],
    "post-submission-file-ops": [
        ["cp", "input/assignment.sh", "work/assignment.sh"]
    ],
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
    local message=""

    eq_one 1 true 'Expected true.' || { score=$((score-1)); message+="Failed to retrun true for good value. "; }
    eq_one 2 false 'Expected false.'|| { score=$((score-1)); message+="Failed to retrun true for good value. "; }

    local json_output='{
        "name": "bash",
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
    local feedback=$3

    local result=$(eq_one $a)
    # BUG in grader!
    [[ $result -ne $expected ]]
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
`

const GOOD_GRADER = `
function test_eq_one() {
    local a=$1
    local expected=$2
    local feedback=$3

    local result=$(eq_one $a)
    # BUG in grader!
    [[ $result -eq $expected ]]
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && main "$@"
`

const SUBMISSION = `
function eq_one() {
    local a=$1

    [[ $a -eq 1 ]]
}
`

/*
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

        // tempDir, err := util.MkDirTemp("regrade-grading-dir-")
        tempDir, inputDir, outputDir, workDir, err := common.PrepTempGradingDir("regrade-grading-dir-")
        if err != nil {
            test.Errorf("Failed to create temp dir for grading: '%v'.", err)
            continue
        }
        defer util.RemoveDirent(tempDir)

        err = util.WriteFile(DUMMY_ASSIGNMENT_CONFIG, filepath.Join(workDir, "assignment.json"))
        if err != nil {
            test.Errorf("Failed to write dummy assignment.json: '%v'.", err)
            continue
        }

        faultyGrader := BASE_GRADER + FAULTY_GRADER
        err = util.WriteFile(faultyGrader, filepath.Join(workDir, "grader.sh"))
        if err != nil {
            test.Errorf("Failed to write faulty grader: '%v'.", err)
            continue
        }

        err = util.WriteFile(SUBMISSION, filepath.Join(inputDir, "assignment.sh"))
        if err != nil {
            test.Errorf("Failed to write submission file: '%v'.", err)
            continue
        }

        result, reject, softError, err := Grade(context.Background(), dummyAssignment, inputDir, "course-student@test.edulinq.org")
        fmt.Fprintf(os.Stderr, "result: '%s'.\nreject: '%s'.\nsoftError: '%s'.\nerr: '%v'.\n",
                util.MustToJSONIndent(result), util.MustToJSONIndent(reject),
                util.MustToJSONIndent(softError), err)

        gradeOptions := GetDefaultGradeOptions()
        gradeOptions.NoDocker = true
        gradeOptions.CheckRejection = false

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
*/

func loadDummyAssignment(sourceDir string, test *testing.T) *model.Assignment {
	dummyCourse := &model.Course{
		ID:          "dummy-course",
		Name:        "Dummy Course",
		Assignments: map[string]*model.Assignment{},
		Source: &util.FileSpec{
			Type: util.FILESPEC_TYPE_PATH,
			Path: filepath.Join(sourceDir, "course.json"),
		},
	}

	err := dummyCourse.Validate()
	if err != nil {
		test.Fatalf("Failed to validate dummy course: '%v'.", err)
	}

	dummyAssignment := &model.Assignment{
		Course:       dummyCourse,
		ID:           "dummy-assignment",
		Name:         "A dummy assignment to test regrades",
		MaxPoints:    1,
		RelSourceDir: "dummy-assignment",
		ImageInfo: docker.ImageInfo{
			StaticFiles: []*util.FileSpec{
				&util.FileSpec{
					Type: util.FILESPEC_TYPE_PATH,
					Path: "grader.sh",
				},
			},
			Invocation: []string{"bash", "./grader.sh"},
			PostSubmissionFileOperations: []*util.FileOperation{
				util.NewFileOperation([]string{"cp", "input/assignment.sh", "work/assignment.sh"}),
			},
			BaseDirFunc: func() (string, string) {
				return sourceDir, sourceDir
			},
		},
	}

	err = dummyAssignment.Validate()
	if err != nil {
		test.Fatalf("Failed to validate dummy assignment: '%v'.", err)
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

	err = db.UpsertUsers(users)
	if err != nil {
		test.Fatalf("Failed to upsert users into the new course: '%v'.", err)
	}

	return dummyAssignment
}

func TestDummyPrep(test *testing.T) {
	db.ResetForTesting()

	// tempDir, err := util.MkDirTemp("regrade-grading-dir-")
	tempDir, inputDir, _, workDir, err := common.PrepTempGradingDir("regrade-grading-dir-")
	if err != nil {
		test.Fatalf("Failed to create temp dir for grading: '%v'.", err)
	}
	defer util.RemoveDirent(tempDir)

	dummyAssignment := loadDummyAssignment(tempDir, test)

	err = util.WriteFile(DUMMY_ASSIGNMENT_CONFIG, filepath.Join(workDir, "assignment.json"))
	if err != nil {
		test.Fatalf("Failed to write dummy assignment.json: '%v'.", err)
	}

	faultyGrader := BASE_GRADER + FAULTY_GRADER
	err = util.WriteFile(faultyGrader, filepath.Join(workDir, "grader.sh"))
	if err != nil {
		test.Fatalf("Failed to write faulty grader: '%v'.", err)
	}

	err = util.WriteFile(SUBMISSION, filepath.Join(inputDir, "assignment.sh"))
	if err != nil {
		test.Fatalf("Failed to write submission file: '%v'.", err)
	}

	gradeOptions := GetDefaultGradeOptions()
	gradeOptions.NoDocker = true
	gradeOptions.CheckRejection = false

	result, reject, softError, err := Grade(context.Background(), dummyAssignment, inputDir, "course-student@test.edulinq.org", "", gradeOptions)
	fmt.Fprintf(os.Stderr, "result: '%s'.\nreject: '%s'.\nsoftError: '%s'.\nerr: '%v'.\n",
		util.MustToJSONIndent(result), util.MustToJSONIndent(reject),
		util.MustToJSONIndent(softError), err)
}

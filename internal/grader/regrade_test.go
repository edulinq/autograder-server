package grader

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

const GOOD_GRADER = `[[ $result -eq $expected ]]`
const FAULTY_GRADER = `[[ $result -ne $expected ]]`

func TestRegradeBase(test *testing.T) {
	defer db.ResetForTesting()

	// A time in the year 3003 which can be used for regrade tests that want a future RegradeAfter time.
	farFutureTime := timestamp.FromMSecs(32614181465000)

	testCases := []struct {
		users              []model.CourseUserReference
		initialSubmissions []string
		waitForCompletion  bool
		numLeft            int
		regradeAfter       *timestamp.Timestamp
		results            map[string]*model.SubmissionHistoryItem
	}{
		// Wait For Completion

		// User With Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
			},
		},

		// User Without Submission
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			nil,
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": nil,
			},
		},

		// Empty Users
		{
			nil,
			nil,
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{},
			nil,
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// All Users, Multiple Submissions
		{
			model.NewAllCourseUserReference(),
			[]string{"course-student@test.edulinq.org", "course-admin@test.edulinq.org"},
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-admin@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
				"course-grader@test.edulinq.org": nil,
				"course-other@test.edulinq.org":  nil,
				"course-owner@test.edulinq.org":  nil,
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
			},
		},

		// Some Users, Multiple Submissions
		{
			[]model.CourseUserReference{"*", "-other", "-owner"},
			[]string{"course-student@test.edulinq.org", "course-grader@test.edulinq.org"},
			true,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{
				"course-admin@test.edulinq.org": nil,
				"course-grader@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-grader@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
			},
		},

		// Regrade Time Before Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			true,
			0,
			timestamp.ZeroPointer(),
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					// Note the full credit came from the original submission with the good grader.
					Score: 10,
				},
			},
		},

		// Regrade Time After Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			true,
			0,
			&farFutureTime,
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					Score:        0,
				},
			},
		},

		// No Wait For Completion

		// User With Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			false,
			1,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// User Without Submission
		{
			[]model.CourseUserReference{"course-admin@test.edulinq.org"},
			nil,
			false,
			1,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Empty Users
		{
			nil,
			nil,
			false,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},
		{
			[]model.CourseUserReference{},
			nil,
			false,
			0,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// All Users, Multiple Submissions
		{
			model.NewAllCourseUserReference(),
			[]string{"course-student@test.edulinq.org", "course-admin@test.edulinq.org"},
			false,
			5,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Some Users, Multiple Submissions
		{
			[]model.CourseUserReference{"*", "-other", "-owner"},
			[]string{"course-student@test.edulinq.org", "course-grader@test.edulinq.org"},
			false,
			3,
			nil,
			map[string]*model.SubmissionHistoryItem{},
		},

		// Regrade Time Before Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			false,
			0,
			timestamp.ZeroPointer(),
			map[string]*model.SubmissionHistoryItem{
				"course-student@test.edulinq.org": &model.SubmissionHistoryItem{
					CourseID:     "course-languages",
					AssignmentID: "bash",
					User:         "course-student@test.edulinq.org",
					MaxPoints:    10,
					// Note the full credit came from the original submission with the good grader.
					Score: 10,
				},
			},
		},

		// Regrade Time After Submission
		{
			[]model.CourseUserReference{"course-student@test.edulinq.org"},
			[]string{"course-student@test.edulinq.org"},
			false,
			1,
			&farFutureTime,
			map[string]*model.SubmissionHistoryItem{},
		},
	}

	gradeOptions := GetDefaultGradeOptions()
	gradeOptions.NoDocker = true
	gradeOptions.CheckRejection = false

	for i, testCase := range testCases {
		db.ResetForTesting()

		// Directory where all the test courses and other materials are located.
		baseDir := config.GetTestdataDir()
		bashSolutionDir := filepath.Join(baseDir, "course-languages", "bash", "test-submissions", "solution")

		testSubmissions, err := GetTestSubmissions(bashSolutionDir, false)
		if err != nil {
			test.Errorf("Case %d: Error getting test submissions in '%s': '%v'.", i, baseDir, err)
			continue
		}

		if len(testSubmissions) != 1 {
			test.Errorf("Case %d: Unexpected number of test submissions. Expected: '1', Actual: '%d' in '%s'.", i, len(testSubmissions), baseDir)
			continue
		}

		for _, user := range testCase.initialSubmissions {
			// Capture the grading start time of the initial submission.
			gradingStartTime, err := makeInitialSubmission(user, testSubmissions[0], gradeOptions)
			if err != nil {
				test.Errorf("Case %d: Failed to make initial submissions for user '%s': '%v'.", i, user, err)
				continue
			}

			expectedGradingInfo, ok := testCase.results[user]
			if !ok {
				continue
			}

			// The regrade start time must match the initial submission time.
			expectedGradingInfo.GradingStartTime = gradingStartTime
		}

		workDir := config.GetWorkDir()
		bashGraderPath := filepath.Join(workDir, "sources", "course-languages", "bash", "grader.sh")
		bashGrader := util.MustReadFile(bashGraderPath)

		// Insert a buggy line in the grader that will cause incorrect scoring.
		bashGrader = strings.Replace(bashGrader, GOOD_GRADER, FAULTY_GRADER, 1)
		err = util.WriteFile(bashGrader, bashGraderPath)
		if err != nil {
			test.Errorf("Case %d: Failed to write faulty grader: '%v'.", i, err)
			continue
		}

		assignment := db.MustGetAssignment("course-languages", "bash")
		options := RegradeOptions{
			GradeOptions: gradeOptions,
			JobOptions: jobmanager.JobOptions{
				WaitForCompletion: testCase.waitForCompletion,
			},
			RawReferences:         testCase.users,
			RegradeAfter:          testCase.regradeAfter,
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

		// Ensure there are not any more regrades running in the background.
		if !testCase.waitForCompletion {
			// Use the same regrade after time to continue the previous job.
			options.RegradeAfter = &result.RegradeAfter
			options.JobOptions.WaitForCompletion = true

			_, numLeft, err = Regrade(assignment, options)
			if err != nil {
				test.Errorf("Case %d: Failed to complete cleanup regrade: '%v'.", i, err)
				continue
			}

			if numLeft != 0 {
				test.Errorf("Case %d: Unexpected number of cleanup regrades remaining. Expected: '0', Actual: '%d'.", i, numLeft)
				continue
			}
		}
	}
}

func TestRegradeSubmissionTime(test *testing.T) {
	earlyTime := timestamp.Zero()
	futureTime := earlyTime + 10
	middleTime := (futureTime + earlyTime) / 2

	testCases := []struct {
		gradingStartTime timestamp.Timestamp
		proxyStartTime   *timestamp.Timestamp
		regradeAfter     timestamp.Timestamp
		expected         bool
	}{
		// Normal Submissions, Submitted After Regrade Threshold
		{earlyTime, nil, earlyTime, false},
		{middleTime, nil, earlyTime, false},

		// Normal Submissions, Submitted Before Regrade Threshold
		{earlyTime, nil, middleTime, true},
		{middleTime, nil, futureTime, true},

		// Proxy Submissions, Submitted After Regrade Threshold
		{earlyTime, &earlyTime, earlyTime, false},
		{earlyTime, &futureTime, earlyTime, false},
		{middleTime, &earlyTime, earlyTime, false},
		{middleTime, &middleTime, earlyTime, false},

		{earlyTime, &middleTime, middleTime, false},
		{earlyTime, &futureTime, middleTime, false},
		{middleTime, &futureTime, futureTime, false},

		// Proxy Submissions, Submitted Before Regrade Threshold
		{earlyTime, &earlyTime, middleTime, true},
		{middleTime, &earlyTime, futureTime, true},
	}

	for i, testCase := range testCases {
		result := isSubmittedBeforeRegradeTime(testCase.gradingStartTime, testCase.proxyStartTime, testCase.regradeAfter)
		if result != testCase.expected {
			test.Errorf("Case %d: Unexpected result. Expected: '%t', Actual: '%t'.", i, testCase.expected, result)
			continue
		}
	}
}

func makeInitialSubmission(user string, testSubmission *TestSubmissionInfo, gradeOptions GradeOptions) (timestamp.Timestamp, error) {
	initialMessage := fmt.Sprintf("Submission '%s': ", testSubmission.ID)
	result, reject, softError, err := Grade(context.Background(), testSubmission.Assignment, testSubmission.Dir, user, TEST_MESSAGE, gradeOptions)
	if err != nil {
		message := ""
		if result != nil {
			message += fmt.Sprintf("\n--- stdout ---\n%v\n--------------", result.Stdout)
			message += fmt.Sprintf("\n--- stderr ---\n%v\n--------------", result.Stderr)
		}

		return timestamp.Zero(), fmt.Errorf("%sFailed to grade assignment: '%v'.%s", initialMessage, err, message)
	}

	if reject != nil {
		return timestamp.Zero(), fmt.Errorf("%sSubmission was rejected: '%s'.", initialMessage, reject.String())
	}

	if softError != "" {
		return timestamp.Zero(), fmt.Errorf("%sSubmission got a soft error: '%s'.", initialMessage, softError)
	}

	if !result.Info.Equals(*testSubmission.TestSubmission.GradingInfo, !testSubmission.TestSubmission.IgnoreMessages) {
		return timestamp.Zero(), fmt.Errorf("%sActual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
			initialMessage, util.MustToJSONIndent(result.Info), util.MustToJSONIndent(testSubmission.TestSubmission.GradingInfo))
	}

	return result.Info.GradingStartTime, nil
}

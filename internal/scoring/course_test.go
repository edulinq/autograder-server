package scoring

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestCourseScoringBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	// Set the assignment's LMS ID.
	course.Assignments["hw0"].LMSID = "001"

	runAndtestResult(test, course)
}

// Make sure a student has no submissions.
func TestCourseScoringEmptyStudent(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	// Set the assignment's LMS ID.
	course.Assignments["hw0"].LMSID = "001"

	// Change a user with no submissions into a student.
	user, err := db.GetServerUser("course-other@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get user: '%v'.", err)
	}

	user.CourseInfo[db.TEST_COURSE_ID].Role = model.CourseRoleStudent
	err = db.UpsertUser(user)
	if err != nil {
		test.Fatalf("Failed to save user: '%v'.", err)
	}

	runAndtestResult(test, course)
}

func runAndtestResult(test *testing.T, course *model.Course) {
	actual, err := FullCourseScoringAndUpload(course, true)
	if err != nil {
		test.Fatalf("Course score upload (dryrun) failed: '%v'.", err)
	}

	// Clear the upload time.
	actual["hw0"]["course-student@test.edulinq.org"].UploadTime = timestamp.Zero()

	if !reflect.DeepEqual(expected, actual) {
		test.Fatalf("Result not as expected (dry run: %v). Expected: '%s', Actual: '%s'.",
			true, util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
	}

	actual, err = FullCourseScoringAndUpload(course, false)
	if err != nil {
		test.Fatalf("Course score upload failed: '%v'.", err)
	}

	// Clear the upload time.
	actual["hw0"]["course-student@test.edulinq.org"].UploadTime = timestamp.Zero()

	if !reflect.DeepEqual(expected, actual) {
		test.Fatalf("Result not as expected (dry run: %v). Expected: '%s', Actual: '%s'.",
			false, util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
	}
}

var expected map[string]map[string]*model.ScoringInfo = map[string]map[string]*model.ScoringInfo{
	"hw0": map[string]*model.ScoringInfo{
		"course-student@test.edulinq.org": &model.ScoringInfo{
			ID:                      "course101::hw0::course-student@test.edulinq.org::1697406272",
			SubmissionTime:          timestamp.FromMSecs(1697406273000),
			UploadTime:              timestamp.Zero(), // Use a zero time.
			RawScore:                2,
			Score:                   2,
			Lock:                    false,
			LateDayUsage:            0,
			NumDaysLate:             0,
			Reject:                  false,
			AutograderStructVersion: model.SCORING_INFO_STRUCT_VERSION,
		},
	},
}

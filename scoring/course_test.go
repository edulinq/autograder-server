package scoring

import (
    "testing"

    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model"
)

func TestCourseScoringBase(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    course := db.MustGetTestCourse();

    // Set the assignment's LMS ID.
    course.Assignments["hw0"].LMSID = "001";

    err := FullCourseScoringAndUpload(course, true);
    if (err != nil) {
        test.Fatalf("Course score upload (dryrun) failed: '%v'.", err);
    }

    err = FullCourseScoringAndUpload(course, false);
    if (err != nil) {
        test.Fatalf("Course score upload failed: '%v'.", err);
    }
}

// Make sure a student has no submissions.
func TestCourseScoringEmptyStudent(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    course := db.MustGetTestCourse();

    // Set the assignment's LMS ID.
    course.Assignments["hw0"].LMSID = "001";

    // Change a user with no submissions into a student.
    user, err := db.GetUser(course, "other@test.com");
    if (err != nil) {
        test.Fatalf("Failed to get user: '%v'.", err);
    }

    user.Role = model.RoleStudent;
    err = db.SaveUsers(course, map[string]*model.User{user.Email: user});
    if (err != nil) {
        test.Fatalf("Failed to save user: '%v'.", err);
    }

    err = FullCourseScoringAndUpload(course, true);
    if (err != nil) {
        test.Fatalf("Course score upload (dryrun) failed: '%v'.", err);
    }

    err = FullCourseScoringAndUpload(course, false);
    if (err != nil) {
        test.Fatalf("Course score upload failed: '%v'.", err);
    }
}

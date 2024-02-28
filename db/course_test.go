package db

import (
    "testing"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/model"
)

// Update a course from a path source.
func (this *DBTests) DBTestCourseUpdateCourseFromSourceBase(test *testing.T) {
    defer ResetForTesting();
    ResetForTesting();

    course := MustGetTestCourse();

    count := countUsers(test, course);
    if (count != 5) {
        test.Fatalf("Unexpected pre-remove user count. Expected 5, found %d.", count);
    }

    removed, err := RemoveUser(course, "student@test.com");
    if (err != nil) {
        test.Fatalf("Error when removing the user: '%v'.", err);
    }

    if (!removed) {
        test.Fatalf("User was not removed.")
    }

    count = countUsers(test, course);
    if (count != 4) {
        test.Fatalf("Unexpected post-remove user count. Expected 4, found %d.", count);
    }

    newCourse, updated, err := UpdateCourseFromSource(course);
    if (err != nil) {
        test.Fatalf("Failed to update course: '%v'.", err);
    }

    if (!updated) {
        test.Fatalf("Course did not update.");
    }

    count = countUsers(test, newCourse);
    if (count != 5) {
        test.Fatalf("Unexpected post-update user count. Expected 5, found %d.", count);
    }
}

// Set the course's source to nil and then update.
// This will cause the course to skip updating.
func (this *DBTests) DBTestCourseUpdateCourseFromSourceSkip(test *testing.T) {
    defer ResetForTesting();
    ResetForTesting();

    course := MustGetTestCourse();

    count := countUsers(test, course);
    if (count != 5) {
        test.Fatalf("Unexpected pre-remove user count. Expected 5, found %d.", count);
    }

    removed, err := RemoveUser(course, "student@test.com");
    if (err != nil) {
        test.Fatalf("Error when removing the user: '%v'.", err);
    }

    if (!removed) {
        test.Fatalf("User was not removed.")
    }

    count = countUsers(test, course);
    if (count != 4) {
        test.Fatalf("Unexpected post-remove user count. Expected 4, found %d.", count);
    }

    // Set the source to nil.
    course.Source = common.GetNilFileSpec();
    err = SaveCourse(course);
    if (err != nil) {
        test.Fatalf("Failed to save course: '%v'.", err);
    }

    _, updated, err := UpdateCourseFromSource(course);
    if (err != nil) {
        test.Fatalf("Failed to update course: '%v'.", err);
    }

    if (updated) {
        test.Fatalf("Course was (incorrectly) updated.");
    }

    // We canactually use the old course to still get a count.
    count = countUsers(test, course);
    if (count != 4) {
        test.Fatalf("Unexpected post-update user count. Expected 4, found %d.", count);
    }
}

func countUsers(test *testing.T, course *model.Course) int {
    users, err := GetUsers(course);
    if (err != nil) {
        test.Fatalf("Failed to get users: '%v'.", err);
    }

    return len(users);
}

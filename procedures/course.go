// Common high-level procedures that can be called on by the server or the api.
package procedures

import (
    "errors"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/lms/lmssync"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/task"
)

// Update a live course.
// Keep in sync with task.updateCourse().
func UpdateCourse(course *model.Course, startTasks bool) (bool, error) {
    var errs error;

    // Stop any existing tasks.
    task.StopCourse(course.GetID());

    // Update the course.
    newCourse, updated, err := db.UpdateCourseFromSource(course);
    if (err != nil) {
        // On failure, still try and work with the old course.
        log.Error("Failed to update course.", err, course);
        errs = errors.Join(errs, err);
    } else {
        // On success, use the new course.
        course = newCourse;
    }

    // Sync the course.
    _, err = lmssync.SyncLMS(course, false, true);
    if (err != nil) {
        log.Error("Failed to sync course with LMS.", err, course);
        errs = errors.Join(errs, err);
    }

    // Build images.
    _, buildErrs := course.BuildAssignmentImages(false, false, docker.NewBuildOptions());
    for imageName, err := range buildErrs {
        log.Error("Failed to build image.", err, course, log.NewAttr("image", imageName));
        errs = errors.Join(errs, err);
    }

    // Schedule tasks.
    if (startTasks) {
        for _, courseTask := range course.GetTasks() {
            err = task.Schedule(course, courseTask);
            if (err != nil) {
                log.Error("Failed to schedule task.", err, course, log.NewAttr("task", courseTask.String()));
                errs = errors.Join(errs, err);
            }
        }
    }

    return updated, errs;
}

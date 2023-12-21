// Common high-level procedures that can be called on by the server or the api.
package procedures

import (
    "errors"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/task"
)

func UpdateCourse(course *model.Course) (bool, error) {
    var errs error;

    // Stop any existing tasks.
    task.StopCourse(course.GetID());

    // Update the course.
    newCourse, updated, err := db.UpdateCourseFromSource(course);
    if (err != nil) {
        // On failure, still try and work with the old course.
        log.Error().Err(err).Str("course-id", course.GetID()).Msg("Failed to update course.");
        errs = errors.Join(errs, err);
    } else {
        // On success, use the new course.
        course = newCourse;
    }

    // Build images.
    _, buildErrs := course.BuildAssignmentImages(false, false, docker.NewBuildOptions());
    for imageName, err := range buildErrs {
        log.Error().Err(err).Str("course-id", course.GetID()).Str("image", imageName).Msg("Failed to build image.");
        errs = errors.Join(errs, err);
    }

    // Schedule tasks.
    for _, courseTask := range course.GetTasks() {
        err = task.Schedule(course, courseTask);
        if (err != nil) {
            log.Error().Err(err).Str("course-id", course.GetID()).Str("task", courseTask.String()).Msg("Failed to schedule task.");
            errs = errors.Join(errs, err);
        }
    }

    return updated, errs;
}

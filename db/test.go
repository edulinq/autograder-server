package db

import (
    "path/filepath"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/model"
)

const TEST_COURSE_ID = "COURSE101";

// Load the test course into the database.
// Testing mode should already be enabled.
func MustLoadTestCourse() {
    MustLoadCourse(filepath.Join(config.COURSES_ROOT.Get(), TEST_COURSE_ID, types.COURSE_CONFIG_FILENAME));
}

func MustGetTestCourse() model.Course {
    return MustGetCourse(TEST_COURSE_ID);
}

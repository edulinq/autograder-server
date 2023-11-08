package disk

import (
    "fmt"
    "path/filepath"

    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/util"
)

const DISK_DB_ASSIGNMENTS_DIR = "assignments";

func (this *backend) SaveAssignment(assignment *types.Assignment) error {
    return this.saveAssignmentLock(assignment, true)
}

func (this *backend) saveAssignmentLock(assignment *types.Assignment, acquireLock bool) error {
    if (acquireLock) {
        this.lock.Lock();
        defer this.lock.Unlock();
    }

    util.MkDir(this.getAssignmentDir(assignment));

    err := util.ToJSONFileIndent(assignment, this.getAssignmentPath(assignment));
    if (err != nil) {
        return fmt.Errorf("Failed to save assignment '%s': '%v'.", assignment.FullID(), err);
    }

    return nil;
}

func (this *backend) getAssignmentDir(assignment *types.Assignment) string {
    return filepath.Join(this.getCourseDir(assignment.Course), DISK_DB_ASSIGNMENTS_DIR, assignment.GetID());
}

func (this *backend) getAssignmentPath(assignment *types.Assignment) string {
    return filepath.Join(this.getAssignmentDir(assignment), types.ASSIGNMENT_CONFIG_FILENAME);
}

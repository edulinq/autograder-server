package disk

import (
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db/types"
)

func (this *backend) saveSubmissionsLock(course *types.Course, submissions []*artifact.GradingResult, acquireLock bool) error {
    // TEST
    return nil;
}

func (this *backend) SaveSubmissions(course *types.Course, submissions []*artifact.GradingResult) error {
    // TEST
    return nil;
}

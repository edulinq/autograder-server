package model2

import (
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/lms/adapter"
    "github.com/eriq-augustine/autograder/usr"
)

const COURSE_CONFIG_FILENAME = "course.json"

type Course interface {
    GetID() string
    GetName() string
    GetSourceDir() string
    GetLMSAdapter() *adapter.LMSAdapter
    GetAssignment(id string) Assignment;
    GetAssignments() map[string]Assignment;
    GetSortedAssignments() []Assignment
    GetAssignmentLMSIDs() ([]string, []string)

    GetUser(email string) (*usr.User, error);
    GetUsers() (map[string]*usr.User, error)
    // TODO(eriq): Save a single user.
    SaveUsers(users map[string]*usr.User) error;
    AddUser(user *usr.User, merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error);
    SyncNewUsers(newUsers map[string]*usr.User, merge bool, dryRun bool, sendEmails bool) (*usr.UserSyncResult, error);

    Activate() error;
    BuildAssignmentImages(force bool, quick bool, options *docker.BuildOptions) ([]string, map[string]error);
    GetCacheDir() string;

    SetSourcePathForTesting(sourcePath string) string;
}

package model

import (
    "github.com/eriq-augustine/autograder/docker"
)

type Course interface {
    GetID() string
    GetName() string
    GetSourceDir() string
    GetLMSAdapter() *LMSAdapter
    HasAssignment(id string) bool;
    GetAssignment(id string) Assignment;
    GetAssignments() map[string]Assignment;
    GetSortedAssignments() []Assignment;
    GetAssignmentLMSIDs() ([]string, []string);

    GetTasks() []ScheduledTask;

    BuildAssignmentImages(force bool, quick bool, options *docker.BuildOptions) ([]string, map[string]error);
    GetCacheDir() string;
}

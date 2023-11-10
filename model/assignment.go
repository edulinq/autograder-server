package model

import (
    "strings"

    "github.com/eriq-augustine/autograder/docker"
)

type Assignment interface {
    GetCourse() Course;
    GetID() string;
    GetSortID() string;
    FullID() string;
    GetName() string;
    GetLMSID() string;
    GetLatePolicy() LateGradingPolicy;
    ImageName() string;
    GetImageInfo() *docker.ImageInfo;
    GetSourceDir() string;

    BuildImageQuick() error;
    BuildImage(force bool, quick bool, options *docker.BuildOptions) error;
}

func CompareAssignments(a Assignment, b Assignment) int {
    if ((a == nil) && (b == nil)) {
        return 0;
    }

    // Favor non-nil over nil.
    if (a == nil) {
        return 1;
    } else if (b == nil) {
        return -1;
    }

    aSortID := a.GetSortID();
    bSortID := b.GetSortID();

    // If both don't have sort keys, just use the IDs.
    if ((aSortID == "") && (bSortID == "")) {
        return strings.Compare(a.GetID(), b.GetID());
    }


    // Favor assignments with a sort key over those without.
    if (aSortID == "") {
        return 1;
    } else if (bSortID == "") {
        return -1;
    }

    // Both assignments have a sort key, use that for comparison.
    return strings.Compare(aSortID, bSortID);
}

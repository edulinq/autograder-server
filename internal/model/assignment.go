package model

import (
    "fmt"
    "path/filepath"
    "strings"
    "sync"

    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/docker"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/util"
)

const DEFAULT_SUBMISSIONS_DIR = "_submissions"

const FILE_CACHE_FILENAME = "filecache.json"
const CACHE_FILENAME = "cache.json"

type Assignment struct {
    ID string `json:"id"`
    Name string `json:"name"`
    SortID string `json:"sort-id,omitempty"`

    DueDate common.Timestamp `json:"due-date,omitempty"`
    MaxPoints float64 `json:"max-points,omitempty"`

    LMSID string `json:"lms-id,omitempty"`
    LatePolicy *LateGradingPolicy `json:"late-policy,omitempty"`

    SubmissionLimit *SubmissionLimitInfo `json:"submission-limit,omitempty"`

    docker.ImageInfo

    // Ignore these fields in JSON.
    RelSourceDir string `json:"_rel_source-dir"`
    Course *Course `json:"-"`

    imageLock *sync.Mutex `json:"-"`
}

func (this *Assignment) GetID() string {
    return this.ID;
}

func (this *Assignment) LogValue() []*log.Attr {
    return []*log.Attr{
        log.NewCourseAttr(this.Course.ID),
        log.NewAssignmentAttr(this.ID),
    };
}

func (this *Assignment) GetSortID() string {
    return this.SortID;
}

func (this *Assignment) FullID() string {
    return fmt.Sprintf("%s-%s", this.Course.GetID(), this.ID);
}

func (this *Assignment) GetCourse() *Course {
    return this.Course;
}

func (this *Assignment) GetName() string {
    return this.Name;
}

func (this *Assignment) GetLMSID() string {
    return this.LMSID;
}

func (this *Assignment) GetLatePolicy() LateGradingPolicy {
    return *this.LatePolicy;
}

func (this *Assignment) GetSubmissionLimit() *SubmissionLimitInfo {
    return this.SubmissionLimit;
}

func (this *Assignment) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.Course.GetID(), this.ID));
}

func (this *Assignment) GetImageInfo() *docker.ImageInfo {
    return &this.ImageInfo;
}

func (this *Assignment) GetSourceDir() string {
    return filepath.Join(this.Course.GetBaseSourceDir(), this.RelSourceDir);
}

// Ensure that the assignment is formatted correctly.
// Missing optional components will be defaulted correctly.
func (this *Assignment) Validate() error {
    if (this.Course == nil) {
        return fmt.Errorf("No course found for assignment.")
    }

    var err error;
    this.ID, err = common.ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    err = this.DueDate.Validate();
    if (err != nil) {
        return fmt.Errorf("Due date is not a valid timestamp: '%w'.", err);
    }

    if (this.MaxPoints < 0.0) {
        return fmt.Errorf("Max points cannot be negative: %f.", this.MaxPoints);
    }

    this.imageLock = &sync.Mutex{};

    // Inherit submission limit from course or leave nil.
    if ((this.SubmissionLimit == nil) && (this.Course.SubmissionLimit != nil)) {
        this.SubmissionLimit = this.Course.SubmissionLimit;
    }

    if (this.SubmissionLimit != nil) {
        err = this.SubmissionLimit.Validate();
        if (err != nil) {
            return fmt.Errorf("Failed to validate submission limit: '%w'.", err);
        }
    }

    // Inherit late policy from course or default to empty.
    if (this.LatePolicy == nil) {
        if (this.Course.LatePolicy != nil) {
            this.LatePolicy = this.Course.LatePolicy;
        } else {
            this.LatePolicy = &LateGradingPolicy{};
        }
    }

    err = this.LatePolicy.Validate();
    if (err != nil) {
        return fmt.Errorf("Failed to validate late policy: '%w'.", err);
    }

    if (this.RelSourceDir == "") {
        return fmt.Errorf("Relative source dir must not be empty.")
    }

    this.ImageInfo.Name = this.ImageName();
    this.ImageInfo.BaseDir = this.GetSourceDir();

    err = this.ImageInfo.Validate();
    if (err != nil) {
        return fmt.Errorf("Failed to validate docker information: '%w'.", err);
    }

    return nil;
}

func (this *Assignment) GetCacheDir() string {
    dir := filepath.Join(this.Course.GetCacheDir(), "assignment_" + this.ID);
    util.MkDir(dir);
    return dir;
}

func (this *Assignment) GetCachePath() string {
    return filepath.Join(this.GetCacheDir(), CACHE_FILENAME);
}

func (this *Assignment) GetFileCachePath() string {
    return filepath.Join(this.GetCacheDir(), FILE_CACHE_FILENAME);
}

func (this *Assignment) GetImageLock() *sync.Mutex {
    return this.imageLock;
}

func CompareAssignments(a *Assignment, b *Assignment) int {
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

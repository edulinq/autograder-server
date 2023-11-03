package db

import (
    "fmt"
    "path/filepath"
    "strings"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model2"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"
const DEFAULT_SUBMISSIONS_DIR = "_submissions"

const FILE_CACHE_FILENAME = "filecache.json"
const CACHE_FILENAME = "cache.json"

type Assignment struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`
    SortID string `json:"sort-id"`

    LMSID string `json:"lms-id",omitempty`
    LatePolicy model2.LateGradingPolicy `json:"late-policy,omitempty"`

    docker.ImageInfo

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Course *Course `json:"-"`

    dockerLock *sync.Mutex `json:"-"`
}

// Load an assignment config from a given JSON path.
// If the course config is nil, search all parent directories for the course config.
func LoadAssignmentConfig(path string, course *Course) (*Assignment, error) {
    var assignment Assignment;
    err := util.JSONFromFile(path, &assignment);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    assignment.SourcePath = util.ShouldAbs(path);

    if (course == nil) {
        course, err = loadParentCourseConfig(filepath.Dir(path));
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config for '%s': '%w'.", path, err);
        }
    }
    assignment.Course = course;

    err = assignment.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate assignment config (%s): '%w'.", path, err);
    }

    otherAssignment := course.GetAssignment(assignment.GetID());
    if (otherAssignment != nil) {
        return nil, fmt.Errorf(
                "Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
                assignment.GetID(), otherAssignment.GetSourceDir(), assignment.GetSourceDir());
    }
    course.Assignments[assignment.GetID()] = &assignment;

    return &assignment, nil;
}

func MustLoadAssignmentConfig(path string) *Assignment {
    assignment, err := LoadAssignmentConfig(path, nil);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load assignment config.");
    }

    return assignment;
}

func (this *Assignment) GetID() string {
    return this.ID;
}

func (this *Assignment) GetSortID() string {
    return this.SortID;
}

func (this *Assignment) FullID() string {
    return fmt.Sprintf("%s-%s", this.Course.GetID(), this.ID);
}

func (this *Assignment) GetCourse() model2.Course {
    return this.Course;
}

func (this *Assignment) GetName() string {
    return this.DisplayName;
}

func (this *Assignment) GetLMSID() string {
    return this.LMSID;
}

func (this *Assignment) GetLatePolicy() model2.LateGradingPolicy {
    return this.LatePolicy;
}

func (this *Assignment) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.Course.GetID(), this.ID));
}

func (this *Assignment) GetImageInfo() *docker.ImageInfo {
    return &this.ImageInfo;
}

func (this *Assignment) GetSourceDir() string {
    return filepath.Dir(this.SourcePath);
}

// Ensure that the assignment is formatted correctly.
// Missing optional components will be defaulted correctly.
func (this *Assignment) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = common.ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    this.dockerLock = &sync.Mutex{};

    err = this.LatePolicy.Validate();
    if (err != nil) {
        return fmt.Errorf("Failed to validate late policy: '%w'.", err);
    }

    if (this.PreStaticDockerCommands == nil) {
        this.PreStaticDockerCommands = make([]string, 0);
    }

    if (this.PostStaticDockerCommands == nil) {
        this.PostStaticDockerCommands = make([]string, 0);
    }

    if (this.StaticFiles == nil) {
        this.StaticFiles = make([]common.FileSpec, 0);
    }

    for _, staticFile := range this.StaticFiles {
        if (staticFile.IsAbs()) {
            return fmt.Errorf("All static file paths must be relative (to the assignment config file), found: '%s'.", staticFile);
        }
    }

    if (this.PreStaticFileOperations == nil) {
        this.PreStaticFileOperations = make([][]string, 0);
    }

    if (this.PostStaticFileOperations == nil) {
        this.PostStaticFileOperations = make([][]string, 0);
    }

    if (this.PostSubmissionFileOperations == nil) {
        this.PostSubmissionFileOperations = make([][]string, 0);
    }

    if (this.SourcePath == "") {
        return fmt.Errorf("Source path must not be empty.")
    }

    if (this.Course == nil) {
        return fmt.Errorf("No course found for assignment.")
    }

    if ((this.Image == "") && ((this.Invocation == nil) || (len(this.Invocation) == 0))) {
        return fmt.Errorf("Assignment image and invocation cannot both be empty.");
    }

    this.ImageInfo.Name = this.ImageName();
    this.ImageInfo.BaseDir = filepath.Dir(this.SourcePath);

    return nil;
}

func (this *Assignment) GetUsers() (map[string]*usr.User, error) {
    users, err := this.Course.GetUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to get users for assignment '%s': '%w'.", this.FullID(), err);
    }

    return users, nil;
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

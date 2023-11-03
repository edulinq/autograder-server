package db

import (
    "fmt"
    "path/filepath"
    "slices"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/lms/adapter"
    "github.com/eriq-augustine/autograder/model2"
    "github.com/eriq-augustine/autograder/task"
    "github.com/eriq-augustine/autograder/util"
)

const USERS_FILENAME = "users.json"

type Course struct {
    // Required fields.
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    LMSAdapter *adapter.LMSAdapter `json:"lms,omitempty"`

    Backup []*task.BackupTask `json:"backup,omitempty"`
    Report []*task.ReportTask `json:"report,omitempty"`
    ScoringUpload []*task.ScoringUploadTask `json:"scoring-upload,omitempty"`

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Assignments map[string]model2.Assignment `json:"-"`

    tasks []model2.ScheduledCourseTask `json:"-"`
}

func (this *Course) GetID() string {
    return this.ID;
}

func (this *Course) GetName() string {
    return this.DisplayName;
}

func (this *Course) GetSourceDir() string {
    return filepath.Dir(this.SourcePath);
}

func (this *Course) SetSourcePathForTesting(sourcePath string) string {
    oldPath := this.SourcePath;
    this.SourcePath = sourcePath;
    return oldPath;
}

func (this *Course) GetLMSAdapter() *adapter.LMSAdapter {
    return this.LMSAdapter;
}

func (this *Course) GetAssignmentLMSIDs() ([]string, []string) {
    lmsIDs := make([]string, 0, len(this.Assignments));
    assignmentIDs := make([]string, 0, len(this.Assignments));

    for _, assignment := range this.Assignments {
        lmsIDs = append(lmsIDs, assignment.GetLMSID());
        assignmentIDs = append(assignmentIDs, assignment.GetLMSID());
    }

    return lmsIDs, assignmentIDs;
}

func LoadCourseConfig(path string) (*Course, error) {
    var config Course;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    config.SourcePath = util.ShouldAbs(path);

    config.Assignments = make(map[string]model2.Assignment);

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", path, err);
    }

    return &config, nil;
}

func MustLoadCourseConfig(path string) *Course {
    config, err := LoadCourseDirectory(path);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load course config.");
    }

    return config;
}

// Load the course (with its JSON config) and all assignments (JSON configs) recursivley in a directory.
// The path should point to the course config,
// and the directory that path lives in will be searched for assignment configs.
func LoadCourseDirectory(courseConfigPath string) (*Course, error) {
    courseConfig, err := LoadCourseConfig(courseConfigPath);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config at '%s': '%w'.", courseConfigPath, err);
    }

    courseDir := filepath.Dir(courseConfigPath);

    assignmentPaths, err := util.FindFiles(ASSIGNMENT_CONFIG_FILENAME, courseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for assignment configs in '%s': '%w'.", courseDir, err);
    }

    for _, assignmentPath := range assignmentPaths {
        _, err := LoadAssignmentConfig(assignmentPath, courseConfig);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load assignment config '%s': '%w'.", assignmentPath, err);
        }
    }

    return courseConfig, nil;
}

// Ensure this course makes sense.
func (this *Course) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = common.ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    if (this.LMSAdapter != nil) {
        err = this.LMSAdapter.Validate(this);
        if (err != nil) {
            return err;
        }
    }

    // Register tasks.
    for _, task := range this.Backup {
        this.tasks = append(this.tasks, task);
    }

    for _, task := range this.Report {
        this.tasks = append(this.tasks, task);
    }

    for _, task := range this.ScoringUpload {
        this.tasks = append(this.tasks, task);
    }

    // Validate tasks.
    for _, task := range this.tasks {
        err = task.Validate(this);
        if (err != nil) {
            return err;
        }
    }

    return nil;
}

// TODO(eriq): After DBs, the concept of activation will move to tasks.
// Start any scheduled tasks or informal tasks associated with this course.
func (this *Course) Activate() error {
    // Schedule tasks.
    for _, task := range this.tasks {
        task.Schedule();
    }

    // Build images.
    go this.BuildAssignmentImages(false, false, docker.NewBuildOptions());

    return nil;
}

// Returns: (successfull image names, map[imagename]error).
func (this *Course) BuildAssignmentImages(force bool, quick bool, options *docker.BuildOptions) ([]string, map[string]error) {
    goodImageNames := make([]string, 0, len(this.Assignments));
    errors := make(map[string]error);

    for _, assignment := range this.Assignments {
        err := assignment.BuildImage(force, quick, options);
        if (err != nil) {
            log.Error().Err(err).Str("course", this.ID).Str("assignment", assignment.GetID()).
                    Msg("Failed to build assignment docker image.");
            errors[assignment.ImageName()] = err;
        } else {
            goodImageNames = append(goodImageNames, assignment.ImageName());
        }
    }

    return goodImageNames, errors;
}

func (this *Course) GetCacheDir() string {
    return filepath.Join(config.WORK_DIR.Get(), common.CACHE_DIRNAME, "course_" + this.ID);
}

// Check this directory and all parent directories for a course config file.
func loadParentCourseConfig(basepath string) (*Course, error) {
    configPath := util.SearchParents(basepath, model2.COURSE_CONFIG_FILENAME);
    if (configPath == "") {
        return nil, fmt.Errorf("Could not locate course config.");
    }

    return LoadCourseConfig(configPath);
}

func (this *Course) GetAssignment(id string) model2.Assignment {
    return this.Assignments[id];
}

func (this *Course) GetAssignments() map[string]model2.Assignment {
    return this.Assignments;
}

func (this *Course) GetSortedAssignments() []model2.Assignment {
    assignments := make([]model2.Assignment, 0, len(this.Assignments));
    for _, assignment := range this.Assignments {
        assignments = append(assignments, assignment);
    }

    slices.SortFunc(assignments, model2.CompareAssignments);

    return assignments;
}

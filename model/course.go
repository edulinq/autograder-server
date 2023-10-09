package model

import (
    "fmt"
    "path/filepath"
    "slices"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"
const DEFAULT_USERS_FILENAME = "users.json"

type Course struct {
    // Required fields.
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    // Non-required fields that have defaults.
    // Paths are always relative to the course dir.
    UsersFile string `json:"users-file"`

    CanvasInstanceInfo *canvas.CanvasInstanceInfo `json:"canvas,omitempty"`

    Backup []*BackupTask `json:"backup,omitempty"`
    ScoringUpload []*ScoringUploadTask `json:"scoring-upload,omitempty"`

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Assignments map[string]*Assignment `json:"-"`

    tasks []ScheduledCourseTask `json:"-"`
}

func LoadCourseConfig(path string) (*Course, error) {
    var config Course;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    config.SourcePath = util.MustAbs(path);

    if (config.UsersFile == "") {
        config.UsersFile = DEFAULT_USERS_FILENAME;
    }

    config.Assignments = make(map[string]*Assignment);

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
    this.ID, err = ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    // Register tasks.
    for _, task := range this.Backup {
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

// Start any scheduled tasks or informal tasks associated with this course.
func (this *Course) Activate() error {
    // Schedule tasks.
    for _, task := range this.tasks {
        task.Schedule();
    }

    return nil;
}

func (this *Course) GetCacheDir() string {
    return filepath.Join(config.CACHE_DIR.GetString(), "course_" + this.ID);
}

// Check this directory and all parent directories for a course config file.
func loadParentCourseConfig(basepath string) (*Course, error) {
    configPath := util.SearchParents(basepath, COURSE_CONFIG_FILENAME);
    if (configPath == "") {
        return nil, fmt.Errorf("Could not locate course config.");
    }

    return LoadCourseConfig(configPath);
}

func (this *Course) GetUsers() (map[string]*User, error) {
    path := filepath.Join(filepath.Dir(this.SourcePath), this.UsersFile);

    users, err := LoadUsersFile(path);
    if (err != nil) {
        return nil, fmt.Errorf("Faile to deserialize users file '%s': '%w'.", path, err);
    }

    return users, nil;
}

func (this *Course) GetUser(email string) (*User, error) {
    users, err := this.GetUsers();
    if (err != nil) {
        return nil, err;
    }

    user := users[email];
    if (user != nil) {
        return user, nil;
    }

    return nil, nil;
}

func (this *Course) SaveUsersFile(users map[string]*User) error {
    path := filepath.Join(filepath.Dir(this.SourcePath), this.UsersFile);
    return SaveUsersFile(path, users);
}

func (this *Course) GetSortedAssignments() []*Assignment {
    assignments := make([]*Assignment, 0, len(this.Assignments));
    for _, assignment := range this.Assignments {
        assignments = append(assignments, assignment);
    }

    slices.SortFunc(assignments, CompareAssignments);

    return assignments;
}

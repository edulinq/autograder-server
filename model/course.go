package model

import (
	"fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/eriq-augustine/autograder/util"
)

const COURSE_CONFIG_FILENAME = "course.json"
const DEFAULT_USERS_FILENAME = "users.json"

type Course struct {
    // Required fields.
    ID string  `json:"id"`
    DisplayName string `json:"display-name"`

    // Non-required fields that have defaults.
    // If not provided, the directory the config file is in will be used.
    Dir string `json:"dir"`
    // Paths are always relative to Dir.
    SubmissionsDir string `json:"submissions-dir"`
    UsersFile string `json:"users-file"`

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Assignments map[string]*Assignment `json:"-"`
}

const DEFAULT_SUBMISSIONS_DIR = "submissions";
const DEFAULT_USERS_FILE = "users.json";

func LoadCourseConfig(path string) (*Course, error) {
    var config Course;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load course config (%s): '%w'.", path, err);
    }

    config.SourcePath = util.MustAbs(path);

    if (config.Dir == "") {
        config.Dir = filepath.Dir(config.SourcePath);
    }

    if (config.SubmissionsDir == "") {
        config.SubmissionsDir = DEFAULT_SUBMISSIONS_DIR;
    }

    if (config.UsersFile == "") {
        config.UsersFile = DEFAULT_USERS_FILE;
    }

    config.Assignments = make(map[string]*Assignment);

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Could not validate course config (%s): '%w'.", path, err);
    }

    return &config, nil;
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

func (this *Course) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

    return nil;
}

func (this *Course) GetSubmissionsDir() (string, error) {
    path := filepath.Join(this.Dir, this.SubmissionsDir);

    if (util.PathExists(path)) {
        if (!util.IsDir(path)) {
            return "", fmt.Errorf("Submissions dir ('%s') already exists and is not a dir.", path);
        }
    } else {
        err := os.MkdirAll(path, 0755);
        if (err != nil) {
            return "", fmt.Errorf("Failed to make submissions directory ('%s'): '%w'.", path, err);
        }
    }

    return path, nil;
}

func (this *Course) PrepareSubmission(user string) (string, int64, error) {
    submissionsDir, err := this.GetSubmissionsDir();
    if (err != nil) {
        return "", 0, err;
    }

    return this.PrepareSubmissionWithDir(user, submissionsDir);
}

// Prepare a place to hold the student's submission history.
func (this *Course) PrepareSubmissionWithDir(user string, submissionsDir string) (string, int64, error) {
    submissionID := time.Now().Unix();
    var path string;

    for ; ; {
        path = filepath.Join(submissionsDir, user, fmt.Sprintf("%d", submissionID));
        if (!util.PathExists(path)) {
            break;
        }

        // This ID has been used.
        submissionID++;
    }

    err := os.MkdirAll(path, 0755);
    if (err != nil) {
        return "", 0, fmt.Errorf("Failed to make submission directory ('%s'): '%w'.", path, err);
    }

    return path, submissionID, nil;
}

// Check this directory and all parent directories for a course config file.
func loadParentCourseConfig(basepath string) (*Course, error) {
    configPath := util.SearchParents(basepath, COURSE_CONFIG_FILENAME);
    if (configPath == "") {
        return nil, fmt.Errorf("Could not locate course config.");
    }

    return LoadCourseConfig(configPath);
}

func (this *Course) GetUsers() ([]User, error) {
    path := filepath.Join(filepath.Dir(this.SourcePath), DEFAULT_USERS_FILENAME);

    var users []User;
    err := util.JSONFromFile(path, &users);
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

    for _, user := range users {
        if (user.Email == email) {
            return &user, nil;
        }
    }

    return nil, fmt.Errorf("Could not find user '%s' in course '%s'.", email, this.ID);
}

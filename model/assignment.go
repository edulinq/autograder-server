package model

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

const ASSIGNMENT_CONFIG_FILENAME = "assignment.json"
const DEFAULT_SUBMISSIONS_DIR = "_submissions";

const GRADING_INPUT_DIRNAME = "input"
const GRADING_OUTPUT_DIRNAME = "output"
const GRADING_WORK_DIRNAME = "work"

const GRADER_OUTPUT_RESULT_FILENAME = "result.json"
const GRADER_OUTPUT_SUMMARY_FILENAME = "summary.json"

type Assignment struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`
    SortID string `json:"sort-id"`

    CanvasID string `json:"canvas-id",omitempty`
    LatePolicy LateGradingPolicy `json:"late-policy,omitempty"`

    Image string `json:"image,omitempty"`
    PreStaticDockerCommands []string `json:"pre-static-docker-commands,omitempty"`
    PostStaticDockerCommands []string `json:"post-static-docker-commands,omitempty"`

    Invocation []string `json:"invocation,omitempty"`

    StaticFiles []string `json:"static-files,omitempty"`
    PreStaticFileOperations [][]string `json:"pre-static-files-ops,omitempty"`
    PostStaticFileOperations [][]string `json:"post-static-files-ops,omitempty"`

    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`

    // Ignore these fields in JSON.
    SourcePath string `json:"-"`
    Course *Course `json:"-"`
}

// A subset of assignment that is passed to docker images for config.
type DockerAssignmentConfig struct {
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`
}

func (this *Assignment) GetDockerAssignmentConfig() *DockerAssignmentConfig {
    return &DockerAssignmentConfig{
        ID: this.ID,
        DisplayName: this.DisplayName,
        PostSubmissionFileOperations: this.PostSubmissionFileOperations,
    };
}

// Load an assignment config from a given JSON path.
// If the course config is nil, search all parent directories for the course config.
func LoadAssignmentConfig(path string, courseConfig *Course) (*Assignment, error) {
    var config Assignment;
    err := util.JSONFromFile(path, &config);
    if (err != nil) {
        return nil, fmt.Errorf("Could not load assignment config (%s): '%w'.", path, err);
    }

    config.SourcePath = util.MustAbs(path);

    if (courseConfig == nil) {
        courseConfig, err = loadParentCourseConfig(filepath.Dir(path));
        if (err != nil) {
            return nil, fmt.Errorf("Could not load course config for '%s': '%w'.", path, err);
        }
    }
    config.Course = courseConfig;

    err = config.Validate();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate config (%s): '%w'.", path, err);
    }

    otherAssignment := courseConfig.Assignments[config.ID];
    if (otherAssignment != nil) {
        return nil, fmt.Errorf(
                "Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
                config.ID, otherAssignment.SourcePath, config.SourcePath);
    }
    courseConfig.Assignments[config.ID] = &config;

    return &config, nil;
}

func MustLoadAssignmentConfig(path string) *Assignment {
    config, err := LoadAssignmentConfig(path, nil);
    if (err != nil) {
        log.Fatal().Str("path", path).Err(err).Msg("Failed to load assignment config.");
    }

    return config;
}

func (this *Assignment) FullID() string {
    return fmt.Sprintf("%s-%s", this.Course.ID, this.ID);
}

func (this *Assignment) ImageName() string {
    return strings.ToLower(fmt.Sprintf("autograder.%s.%s", this.Course.ID, this.ID));
}

// Ensure that the assignment is formatted correctly.
// Missing optional components will be defaulted correctly.
func (this *Assignment) Validate() error {
    if (this.DisplayName == "") {
        this.DisplayName = this.ID;
    }

    var err error;
    this.ID, err = ValidateID(this.ID);
    if (err != nil) {
        return err;
    }

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
        this.StaticFiles = make([]string, 0);
    }

    for _, staticFile := range this.StaticFiles {
        if (filepath.IsAbs(staticFile)) {
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

    return nil;
}

func (this *Assignment) getSubmissionsDir() (string, error) {
    assignmentDir := filepath.Dir(this.SourcePath);
    path := filepath.Join(assignmentDir, DEFAULT_SUBMISSIONS_DIR);

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

func (this *Assignment) PrepareSubmission(user string) (string, int64, error) {
    submissionsDir, err := this.getSubmissionsDir();
    if (err != nil) {
        return "", 0, err;
    }

    return this.PrepareSubmissionWithDir(user, submissionsDir);
}

// Prepare a place to hold the student's submission history.
func (this *Assignment) PrepareSubmissionWithDir(user string, submissionsDir string) (string, int64, error) {
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

// See getSubmissionFiles().
// Fetches full grading result.
func (this *Assignment) GetSubmissionResults(user string) ([]string, error) {
    return this.getSubmissionFiles(user, GRADER_OUTPUT_RESULT_FILENAME);
}

// See getSubmissionFiles().
// Fetches grading summary.
func (this *Assignment) GetSubmissionSummaries(user string) ([]string, error) {
    return this.getSubmissionFiles(user, GRADER_OUTPUT_SUMMARY_FILENAME);
}

// Get all the paths to the submission files for and assignment and user.
// The results will be sorted in ascending order (first submission first).
// An empty slice indicates that there are no matching submission files.
func (this *Assignment) getSubmissionFiles(user string, filename string) ([]string, error) {
    submissionsDir, err := this.getSubmissionsDir();
    if (err != nil) {
        return nil, err;
    }

    paths := make([]string, 0);

    baseDir := filepath.Join(submissionsDir, user);
    if (!util.PathExists(baseDir)) {
        return paths, nil;
    }

    if (!util.IsDir(baseDir)) {
        return nil, fmt.Errorf("Expected user's submission dir '%s' exists and is not a dir.", baseDir);
    }

    dirents, err := os.ReadDir(baseDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to read dir '%s': '%w'.", baseDir, err);
    }

    for _, dirent := range dirents {
        if (!dirent.IsDir()) {
            continue;
        }

        path := filepath.Join(baseDir, dirent.Name(), GRADING_OUTPUT_DIRNAME, filename);
        if (!util.IsFile(path)) {
            continue;
        }

        paths = append(paths, path);
    }

    return paths, nil;
}

// See getAllRecentSubmissionFiles().
// Fetches full grading result.
func (this *Assignment) GetAllRecentSubmissionResults(users map[string]*User) (map[string]string, error) {
    return this.getAllRecentSubmissionFiles(users, GRADER_OUTPUT_RESULT_FILENAME);
}

// See getAllRecentSubmissionFiles().
// Fetches grading summary.
func (this *Assignment) GetAllRecentSubmissionSummaries(users map[string]*User) (map[string]string, error) {
    return this.getAllRecentSubmissionFiles(users, GRADER_OUTPUT_SUMMARY_FILENAME);
}

// Get all the paths to the most recent submission file for each user for this assignment.
// The returned map will contain an entry for every user (if not nil).
// An empty entry in the map indicates the user has no submissions.
func (this *Assignment) getAllRecentSubmissionFiles(users map[string]*User, filename string) (map[string]string, error) {
    paths := make(map[string]string);

    for email, _ := range users {
        userPaths, err := this.getSubmissionFiles(email, filename);
        if (err != nil) {
            return nil, err;
        }

        if (len(userPaths) == 0) {
            paths[email] = "";
        } else {
            paths[email] = userPaths[len(userPaths) - 1];
        }
    }

    return paths, nil;
}

// Get all the recent submission summaries (via GetAllRecentSubmissionSummaries()),
// and convert them to ScoringInfo structs so they can be properly scored/uploaded.
func (this *Assignment) GetScoringInfo(users map[string]*User, onlyStudents bool) (map[string]*ScoringInfo, error) {
    paths, err := this.GetAllRecentSubmissionSummaries(users);
    if (err != nil) {
        return nil, fmt.Errorf("Unable to load submission summaries: '%w'.", err);
    }

    results := make(map[string]*ScoringInfo, len(paths));

    for username, path := range paths {
        if (path == "") {
            continue;
        }

        if (onlyStudents && (users[username].Role != Student)) {
            continue;
        }

        var summary SubmissionSummary;
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, fmt.Errorf("Unable to load submission summary from path '%s': '%w'.", path, err);
        }

        results[username] = ScoringInfoFromSubmissionSummary(&summary);
    }

    return results, nil;
}

func CompareAssignments(a *Assignment, b *Assignment) int {
    if (a == b) {
        return 0;
    }

    // Favor non-nil over nil.
    if (a == nil) {
        return 1;
    } else if (b == nil) {
        return -1;
    }

    // If both assignments have a sort key, use that for comparison.
    if ((a.SortID != "") && (b.SortID != "")) {
        return strings.Compare(a.SortID, b.SortID);
    }

    // Favor assignments with a sort key over those without.
    if (a.SortID == "") {
        return 1;
    } else if (b.SortID == "") {
        return -1;
    }

    // If both don't have sort keys, just use the IDs.
    return strings.Compare(a.ID, b.ID);
}

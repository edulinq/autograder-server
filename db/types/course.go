package types

import (
    "path/filepath"
    "slices"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
)

type Course struct {
    // Required fields.
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    LMS *model.LMSAdapter `json:"lms,omitempty"`

    Backup []*model.BackupTask `json:"backup,omitempty"`
    Report []*model.ReportTask `json:"report,omitempty"`
    ScoringUpload []*model.ScoringUploadTask `json:"scoring-upload,omitempty"`

    // Ignore these fields in JSON.
    DBID int `json:"-"`
    SourcePath string `json:"-"`

    // TEST - This should go away at some point.
    Assignments map[string]*Assignment `json:"-"`

    tasks []model.ScheduledTask `json:"-"`
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

func (this *Course) GetLMSAdapter() *model.LMSAdapter {
    return this.LMS;
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

    if (this.LMS != nil) {
        err = this.LMS.Validate();
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
    /* TEST
    // Schedule tasks.
    for _, task := range this.tasks {
        task.Schedule();
    }

    // Build images.
    go this.BuildAssignmentImages(false, false, docker.NewBuildOptions());
    */

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

func (this *Course) HasAssignment(id string) bool {
    _, ok := this.Assignments[id];
    return ok;
}

// Get an assignment, or nil if the assignment does not exist.
func (this *Course) GetAssignment(id string) model.Assignment {
    assignment, ok := this.Assignments[id];
    if (!ok) {
        return nil;
    }

    return assignment;
}

func (this *Course) GetAssignments() map[string]model.Assignment {
    assignments := make(map[string]model.Assignment, len(this.Assignments));
    for key, value := range this.Assignments {
        assignments[key] = value;
    }

    return assignments;
}

func (this *Course) GetSortedAssignments() []model.Assignment {
    assignments := make([]model.Assignment, 0, len(this.Assignments));
    for _, assignment := range this.Assignments {
        assignments = append(assignments, assignment);
    }

    slices.SortFunc(assignments, model.CompareAssignments);

    return assignments;
}

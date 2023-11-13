package model

import (
    "path/filepath"
    "slices"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
)

type Course struct {
    // Required fields.
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    LMS *LMSAdapter `json:"lms,omitempty"`

    Backup []*BackupTask `json:"backup,omitempty"`
    Report []*ReportTask `json:"report,omitempty"`
    ScoringUpload []*ScoringUploadTask `json:"scoring-upload,omitempty"`

    // Internal fields the autograder will set.
    SourceDir string `json:"_source-dir"`
    Assignments map[string]*Assignment `json:"-"`
    tasks []ScheduledTask `json:"-"`
}

func (this *Course) GetID() string {
    return this.ID;
}

func (this *Course) GetName() string {
    return this.DisplayName;
}

func (this *Course) GetSourceDir() string {
    return this.SourceDir;
}

func (this *Course) GetLMSAdapter() *LMSAdapter {
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

func (this *Course) GetTasks() []ScheduledTask {
    return this.tasks;
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

// Returns: (successfull image names, map[imagename]error).
func (this *Course) BuildAssignmentImages(force bool, quick bool, options *docker.BuildOptions) ([]string, map[string]error) {
    goodImageNames := make([]string, 0, len(this.Assignments));
    errors := make(map[string]error);

    for _, assignment := range this.Assignments {
        err := docker.BuildImageFromSource(assignment, force, quick, options);
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
func (this *Course) GetAssignment(id string) *Assignment {
    assignment, ok := this.Assignments[id];
    if (!ok) {
        return nil;
    }

    return assignment;
}

func (this *Course) GetAssignments() map[string]*Assignment {
    assignments := make(map[string]*Assignment, len(this.Assignments));
    for key, value := range this.Assignments {
        assignments[key] = value;
    }

    return assignments;
}

func (this *Course) GetSortedAssignments() []*Assignment {
    assignments := make([]*Assignment, 0, len(this.Assignments));
    for _, assignment := range this.Assignments {
        assignments = append(assignments, assignment);
    }

    slices.SortFunc(assignments, CompareAssignments);

    return assignments;
}

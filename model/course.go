package model

import (
    "path/filepath"
    "slices"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model/tasks"
)

const SOURCES_DIRNAME = "sources";

type Course struct {
    // Required fields.
    ID string `json:"id"`
    DisplayName string `json:"display-name"`

    Source common.FileSpec `json:"source"`

    LMS *LMSAdapter `json:"lms,omitempty"`

    Backup []*tasks.BackupTask `json:"backup,omitempty"`
    Report []*tasks.ReportTask `json:"report,omitempty"`
    ScoringUpload []*tasks.ScoringUploadTask `json:"scoring-upload,omitempty"`

    // Internal fields the autograder will set.
    Assignments map[string]*Assignment `json:"-"`
    scheduledTasks []tasks.ScheduledTask `json:"-"`
}

func (this *Course) GetID() string {
    return this.ID;
}

func (this *Course) GetName() string {
    return this.DisplayName;
}

func (this *Course) GetSource() common.FileSpec {
    return this.Source;
}

func (this *Course) GetLMSAdapter() *LMSAdapter {
    return this.LMS;
}

func (this *Course) HasLMSAdapter() bool {
    return (this.LMS != nil);
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

func (this *Course) GetTasks() []tasks.ScheduledTask {
    return this.scheduledTasks;
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
        this.scheduledTasks = append(this.scheduledTasks, task);
    }

    for _, task := range this.Report {
        this.scheduledTasks = append(this.scheduledTasks, task);
    }

    for _, task := range this.ScoringUpload {
        this.scheduledTasks = append(this.scheduledTasks, task);
    }

    // Validate tasks.
    for _, task := range this.scheduledTasks {
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

func GetSourcesDir() string {
    return filepath.Join(config.WORK_DIR.Get(), SOURCES_DIRNAME);
}

func (this *Course) GetBaseSourceDir() string {
    return filepath.Join(GetSourcesDir(), this.GetID());
}

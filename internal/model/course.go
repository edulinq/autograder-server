package model

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	dtasks "github.com/edulinq/autograder/internal/deprecated/model/tasks"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type Course struct {
	// Required fields.
	ID   string `json:"id"`
	Name string `json:"name"`

	Source *util.FileSpec `json:"source"`

	LMS *LMSAdapter `json:"lms,omitempty"`

	// A common late policy that assignments can inherit.
	LatePolicy *LateGradingPolicy `json:"late-policy,omitempty"`

	// A common submission limit that assignments can inherit.
	SubmissionLimit *SubmissionLimitInfo `json:"submission-limit,omitempty"`

	// Deprecated
	Backup        []*dtasks.BackupTask        `json:"backup,omitempty"`
	CourseUpdate  []*dtasks.CourseUpdateTask  `json:"course-update,omitempty"`
	Report        []*dtasks.ReportTask        `json:"report,omitempty"`
	ScoringUpload []*dtasks.ScoringUploadTask `json:"scoring-upload,omitempty"`
	EmailLogs     []*dtasks.EmailLogsTask     `json:"email-logs,omitempty"`

	Tasks []*UserTaskInfo `json:"tasks,omitempty"`

	// Internal fields the autograder will set.
	Assignments map[string]*Assignment `json:"-"`
}

func (this *Course) GetID() string {
	return this.ID
}

func (this *Course) LogValue() []*log.Attr {
	return []*log.Attr{log.NewCourseAttr(this.ID)}
}

func (this *Course) GetName() string {
	return this.Name
}

func (this *Course) GetDisplayName() string {
	if this.Name != "" {
		return this.Name
	}

	return this.ID
}

func (this *Course) GetSource() *util.FileSpec {
	return this.Source
}

func (this *Course) GetLMSAdapter() *LMSAdapter {
	return this.LMS
}

func (this *Course) HasLMSAdapter() bool {
	return (this.LMS != nil)
}

func (this *Course) GetAssignmentLMSIDs() ([]string, []string) {
	lmsIDs := make([]string, 0, len(this.Assignments))
	assignmentIDs := make([]string, 0, len(this.Assignments))

	for _, assignment := range this.Assignments {
		lmsIDs = append(lmsIDs, assignment.GetLMSID())
		assignmentIDs = append(assignmentIDs, assignment.GetID())
	}

	return lmsIDs, assignmentIDs
}

// Ensure this course makes sense.
func (this *Course) Validate() error {
	if this.Name == "" {
		this.Name = this.ID
	}

	var err error
	this.ID, err = common.ValidateID(this.ID)
	if err != nil {
		return err
	}

	if this.LMS != nil {
		err = this.LMS.Validate()
		if err != nil {
			return err
		}
	}

	if this.LatePolicy != nil {
		err = this.LatePolicy.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate late policy: '%w'.", err)
		}
	}

	if this.SubmissionLimit != nil {
		err = this.SubmissionLimit.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate submission limit: '%w'.", err)
		}
	}

	if this.Tasks == nil {
		this.Tasks = make([]*UserTaskInfo, 0)
	}

	// DEP: This is deprecated and will be removed when then the individual task types are.
	this.convertDeprecatedTasks()

	// Validate tasks.
	for i, task := range this.Tasks {
		err = task.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate task at index %d: '%w'.", i, err)
		}
	}

	return nil
}

// Convert deprecated tasks and add them to the task list.
func (this *Course) convertDeprecatedTasks() {
	for _, task := range this.Backup {
		this.Tasks = append(this.Tasks, DeprecatedTaskToStandard(task)...)
	}

	for _, task := range this.CourseUpdate {
		this.Tasks = append(this.Tasks, DeprecatedTaskToStandard(task)...)
	}

	for _, task := range this.Report {
		this.Tasks = append(this.Tasks, DeprecatedTaskToStandard(task)...)
	}

	for _, task := range this.ScoringUpload {
		this.Tasks = append(this.Tasks, DeprecatedTaskToStandard(task)...)
	}

	for _, task := range this.EmailLogs {
		this.Tasks = append(this.Tasks, DeprecatedTaskToStandard(task)...)
	}
}

func (this *Course) AddAssignment(assignment *Assignment) error {
	for _, otherAssignment := range this.Assignments {
		if assignment.GetID() == otherAssignment.GetID() {
			return fmt.Errorf(
				"Found multiple assignments with the same ID ('%s'): ['%s', '%s'].",
				assignment.GetID(), otherAssignment.GetSourceDir(), assignment.GetSourceDir())
		}

		if assignment.GetName() == otherAssignment.GetName() {
			return fmt.Errorf(
				"Found multiple assignments with the same name ('%s'): ['%s', '%s'].",
				assignment.GetName(), otherAssignment.GetID(), assignment.GetID())
		}

		if (assignment.GetLMSID() != "") && (assignment.GetLMSID() == otherAssignment.GetLMSID()) {
			return fmt.Errorf(
				"Found multiple assignments with the same LMS ID ('%s'): ['%s', '%s'].",
				assignment.GetLMSID(), otherAssignment.GetID(), assignment.GetID())
		}
	}

	this.Assignments[assignment.GetID()] = assignment
	return nil
}

// Returns: (successful image names, map[imagename]error).
func (this *Course) BuildAssignmentImages(force bool, quick bool, options *docker.BuildOptions) ([]string, map[string]error) {
	goodImageNames := make([]string, 0, len(this.Assignments))
	errors := make(map[string]error)

	for _, assignment := range this.Assignments {
		err := docker.BuildImageFromSource(assignment, force, quick, options)
		if err != nil {
			log.Error("Failed to build assignment docker image.", err, this, assignment)
			errors[assignment.ImageName()] = err
		} else {
			goodImageNames = append(goodImageNames, assignment.ImageName())
		}
	}

	return goodImageNames, errors
}

// A simpler interface to BuildAssignmentImages().
func (this *Course) BuildAssignmentImagesDefault() ([]string, error) {
	var err error = nil

	imageNames, buildErrs := this.BuildAssignmentImages(false, false, docker.NewBuildOptions())
	for imageName, buildErr := range buildErrs {
		err = errors.Join(err, fmt.Errorf("Failed to build image '%s': '%w'.", imageName, buildErr))
	}

	return imageNames, err
}

func (this *Course) FetchAssignmentTemplateFiles() (map[string][]string, error) {
	result := make(map[string][]string, len(this.Assignments))
	var errs error = nil

	for _, assignment := range this.Assignments {
		relpaths, err := assignment.FetchTemplateFiles()
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			result[assignment.ID] = relpaths
		}
	}

	return result, errs
}

func (this *Course) GetCacheDir() string {
	return filepath.Join(config.GetCacheDir(), "course_"+this.ID)
}

func (this *Course) HasAssignment(id string) bool {
	return (this.GetAssignment(id) != nil)
}

// Get an assignment, or nil if the assignment does not exist.
func (this *Course) GetAssignment(id string) *Assignment {
	id, _ = common.ValidateID(id)

	assignment, ok := this.Assignments[id]
	if !ok {
		return nil
	}

	return assignment
}

func (this *Course) GetAssignments() map[string]*Assignment {
	assignments := make(map[string]*Assignment, len(this.Assignments))
	for key, value := range this.Assignments {
		assignments[key] = value
	}

	return assignments
}

func (this *Course) GetSortedAssignments() []*Assignment {
	assignments := make([]*Assignment, 0, len(this.Assignments))
	for _, assignment := range this.Assignments {
		assignments = append(assignments, assignment)
	}

	slices.SortFunc(assignments, CompareAssignments)

	return assignments
}

func (this *Course) GetBaseSourceDir() string {
	return filepath.Join(config.GetSourcesDir(), this.GetID())
}

func (this *Course) GetTemplatesDir() string {
	return filepath.Join(config.GetTemplatesDir(), this.GetID())
}

func (this *Course) GetSourceConfigPath() string {
	return filepath.Join(this.GetBaseSourceDir(), COURSE_CONFIG_FILENAME)
}

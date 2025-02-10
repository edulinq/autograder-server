package docker

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
)

const (
	DEFAULT_IMAGE = "edulinq/autograder/internal.base"
)

type ImageInfo struct {
	Image                    string   `json:"image,omitempty"`
	PreStaticDockerCommands  []string `json:"pre-static-docker-commands,omitempty"`
	PostStaticDockerCommands []string `json:"post-static-docker-commands,omitempty"`

	Invocation []string `json:"invocation,omitempty"`

	StaticFiles []*common.FileSpec `json:"static-files,omitempty"`

	PreStaticFileOperations  []*common.FileOperation `json:"pre-static-files-ops,omitempty"`
	PostStaticFileOperations []*common.FileOperation `json:"post-static-files-ops,omitempty"`

	PostSubmissionFileOperations []*common.FileOperation `json:"post-submission-files-ops,omitempty"`

	MaxRuntimeSecs int `json:"max-runtime-secs,omitempty"`

	// Fields that are not part of the JSON and are set after deserialization.

	Name string `json:"-"`

	// Dir used for relative paths.
	// Using a func allows for lazy resolution of the base dir.
	BaseDirFunc func() string `json:"-"`
}

// A subset of the image information that is passed to docker images for config during grading.
type GradingConfig struct {
	Name                         string                  `json:"name"`
	PostSubmissionFileOperations []*common.FileOperation `json:"post-submission-files-ops,omitempty"`
}

func (this *ImageInfo) GetGradingConfig() *GradingConfig {
	return &GradingConfig{
		Name:                         this.Name,
		PostSubmissionFileOperations: this.PostSubmissionFileOperations,
	}
}

func (this *ImageInfo) Validate() error {
	if this.Name == "" {
		return fmt.Errorf("Missing name.")
	}

	if this.BaseDirFunc == nil {
		return fmt.Errorf("Missing base dir func.")
	}

	if this.BaseDirFunc() == "" {
		return fmt.Errorf("Missing base dir.")
	}

	if this.Invocation == nil {
		this.Invocation = make([]string, 0)
	}

	if (this.Image == "") && (len(this.Invocation) == 0) {
		return fmt.Errorf("Image and invocation cannot both be empty.")
	}

	if this.Image == "" {
		this.Image = DEFAULT_IMAGE
	}

	if this.PreStaticDockerCommands == nil {
		this.PreStaticDockerCommands = make([]string, 0)
	}

	if this.PostStaticDockerCommands == nil {
		this.PostStaticDockerCommands = make([]string, 0)
	}

	if this.StaticFiles == nil {
		this.StaticFiles = make([]*common.FileSpec, 0)
	}

	for _, staticFile := range this.StaticFiles {
		err := staticFile.Validate()
		if err != nil {
			return fmt.Errorf("Failed to validate static file spec: '%w'.", err)
		}

		if staticFile.IsAbs() {
			return fmt.Errorf("All static file paths must be relative (to the assignment config file), found: '%s'.", staticFile)
		}
	}

	if this.PreStaticFileOperations == nil {
		this.PreStaticFileOperations = make([]*common.FileOperation, 0)
	}

	err := common.ValidateFileOperations(this.PreStaticFileOperations)
	if err != nil {
		return fmt.Errorf("Failed to validate pre-static file operations: '%w'.", err)
	}

	if this.PostStaticFileOperations == nil {
		this.PostStaticFileOperations = make([]*common.FileOperation, 0)
	}

	err = common.ValidateFileOperations(this.PostStaticFileOperations)
	if err != nil {
		return fmt.Errorf("Failed to validate post-static file operations: '%w'.", err)
	}

	if this.PostSubmissionFileOperations == nil {
		this.PostSubmissionFileOperations = make([]*common.FileOperation, 0)
	}

	err = common.ValidateFileOperations(this.PostSubmissionFileOperations)
	if err != nil {
		return fmt.Errorf("Failed to validate post-submission file operations: '%w'.", err)
	}

	if this.MaxRuntimeSecs < 0 {
		return fmt.Errorf("Max runtime seconds must be non-negative, found: %d.", this.MaxRuntimeSecs)
	}

	return nil
}

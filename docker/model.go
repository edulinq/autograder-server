package docker

import (
    "github.com/eriq-augustine/autograder/common"
)

type ImageInfo struct {
    Image string `json:"image,omitempty"`
    PreStaticDockerCommands []string `json:"pre-static-docker-commands,omitempty"`
    PostStaticDockerCommands []string `json:"post-static-docker-commands,omitempty"`

    Invocation []string `json:"invocation,omitempty"`

    StaticFiles []common.FileSpec `json:"static-files,omitempty"`

    PreStaticFileOperations [][]string `json:"pre-static-files-ops,omitempty"`
    PostStaticFileOperations [][]string `json:"post-static-files-ops,omitempty"`

    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`

    // Fields that are not part of the JSON and are set after deserialization.

    Name string `json:"-"`
    // Dir used for relative paths.
    BaseDir string `json:"-"`
}

// A subset of the image information that is passed to docker images for config during grading.
type GradingConfig struct {
    Name string `json:"name"`
    PostSubmissionFileOperations [][]string `json:"post-submission-files-ops,omitempty"`
}

func (this *ImageInfo) GetGradingConfig() *GradingConfig {
    return &GradingConfig{
        Name: this.Name,
        PostSubmissionFileOperations: this.PostSubmissionFileOperations,
    };
}

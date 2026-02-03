package docker

import (
	"context"
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/util"
)

const FILE_PREP_IMAGE = "alpine:3.19"
const FILE_PREP_TIMEOUT_SECS = 60

// Copy directory contents inside a Docker container for isolation.
// Source is mounted read-only, dest is mounted read-write.
// Equivalent to: cp -r /source/* /dest/
// Falls back to host-based copy when Docker is disabled.
func CopyDirContentsInContainer(source string, dest string) error {
	if config.DOCKER_DISABLE.Get() {
		return util.CopyDirContents(source, dest)
	}

	err := EnsureImage(FILE_PREP_IMAGE)
	if err != nil {
		return fmt.Errorf("Failed to ensure file prep image '%s': '%w'", FILE_PREP_IMAGE, err)
	}

	mounts := []MountInfo{
		{
			Source:   util.ShouldAbs(source),
			Target:   "/source",
			ReadOnly: true,
		},
		{
			Source:   util.ShouldAbs(dest),
			Target:   "/dest",
			ReadOnly: false,
		},
	}

	cmd := []string{"sh", "-c", "cp -r /source/. /dest/"}

	stdout, stderr, timeout, _, err := RunContainer(
		context.Background(),
		nil,
		FILE_PREP_IMAGE,
		mounts,
		cmd,
		"fileprep",
		FILE_PREP_TIMEOUT_SECS,
	)

	if err != nil {
		return fmt.Errorf("File prep container failed: %w (stdout: %s, stderr: %s)", err, stdout, stderr)
	}

	if timeout {
		return fmt.Errorf("File prep container timed out (stdout: %s, stderr: %s)", stdout, stderr)
	}

	return nil
}

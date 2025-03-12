package util

import (
	"os"

	"github.com/edulinq/autograder/internal/log"
)

func SymbolicLink(targetPath string, linkPath string) error {
	return os.Symlink(targetPath, linkPath)
}

func MustSymbolicLink(targetPath string, linkPath string) {
	err := SymbolicLink(targetPath, linkPath)
	if err != nil {
		log.Fatal("Failed to created symbolic link.", err, log.NewAttr("target-path", targetPath), log.NewAttr("link-path", linkPath))
	}
}

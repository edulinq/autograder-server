package model

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/util"
)

// Both file specs and file ops for templare files must be valid and local only (relative).
func (this *AssignmentAnalysisOptions) validateTemplateFiles() error {
	var errs error

	for i, spec := range this.TemplateFiles {
		err := spec.ValidateFull(true)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to validate template file spec at index %d ('%s`): '%w'.", i, spec.String(), err))
		}
	}

	for i, op := range this.TemplateFileOps {
		err := op.Validate()
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to validate template file op at index %d ('%s`): '%w'.", i, op.String(), err))
		}
	}

	return errs
}

// Fetch template files using the file specs from the baseDir,
// and then execute any file operations on the target dir.
// Return the relative paths to all the final template files.
func (this *AssignmentAnalysisOptions) FetchTemplateFiles(baseDir string, courseDir string, destDir string) ([]string, error) {
	for i, spec := range this.TemplateFiles {
		err := spec.CopyTarget(baseDir, courseDir, destDir, destDir)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch template file spec at index %d ('%s`): '%w'.", i, spec.String(), err)
		}
	}

	for i, op := range this.TemplateFileOps {
		err := op.Exec(destDir)
		if err != nil {
			return nil, fmt.Errorf("Failed to execute template file op at index %d ('%s`): '%w'.", i, op.String(), err)
		}
	}

	relpaths, err := util.GetAllDirents(destDir, true, true)
	if err != nil {
		return nil, fmt.Errorf("Failed to get paths of final template files: '%w'.", err)
	}

	return relpaths, nil
}

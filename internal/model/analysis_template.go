package model

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/util"
)

// Both file specs and file ops for templare files must be valid and local only (relative).
func (this *AnalysisOptions) validateTemplateFiles() error {
	var errs error

	if this.TemplateFiles == nil {
		this.TemplateFiles = make([]*util.FileSpec, 0)
	}

	for i, spec := range this.TemplateFiles {
		err := spec.ValidateFull(true)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("Failed to validate template file spec at index %d ('%s`): '%w'.", i, spec.String(), err))
		}
	}

	if this.TemplateFileOps == nil {
		this.TemplateFileOps = make([]*util.FileOperation, 0)
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
func (this *AnalysisOptions) FetchTemplateFiles(baseDir string, destDir string) error {
	for i, spec := range this.TemplateFiles {
		err := spec.CopyTarget(baseDir, destDir)
		if err != nil {
			return fmt.Errorf("Failed to fetch template file spec at index %d ('%s`): '%w'.", i, spec.String(), err)
		}
	}

	for i, op := range this.TemplateFileOps {
		err := op.Exec(destDir)
		if err != nil {
			return fmt.Errorf("Failed to execute template file op at index %d ('%s`): '%w'.", i, op.String(), err)
		}
	}

	return nil
}

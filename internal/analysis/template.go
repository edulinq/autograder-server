package analysis

import (
	"fmt"
	"sync"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// A struct for keeping track of assignment template dirs over a pairwise analysis.
type TemplateFileStore struct {
	lock *sync.Mutex
	// {assignmentFullID: tempDir, ...}
	store map[string]string
}

func NewTemplateFileStore() *TemplateFileStore {
	return &TemplateFileStore{
		lock:  &sync.Mutex{},
		store: make(map[string]string),
	}
}

// Get a path to a prepared (via prepSourceFiles()) directory containing all template files for this assignment.
func (this *TemplateFileStore) GetTemplatePath(assignment *model.Assignment) (string, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	id := ""
	if assignment != nil {
		id = assignment.FullID()
	}

	// Check for a cached path.
	path, ok := this.store[id]
	if ok {
		return path, nil
	}

	// Nothing in the cache, get a new one.
	tempDir, err := util.MkDirTemp("analysis-template-file-store-")
	if err != nil {
		return "", fmt.Errorf("Failed to create a temp template file store: '%w'.", err)
	}

	if assignment == nil {
		this.store[id] = tempDir
		return tempDir, nil
	}

	templateDir := assignment.GetTemplatesDir()
	if util.PathExists(templateDir) {
		err = util.CopyDirContents(templateDir, tempDir)
		if err != nil {
			return "", fmt.Errorf("Failed to copy over assignment template files: '%w'.", err)
		}
	}

	_, err = prepSourceFiles(tempDir)
	if err != nil {
		return "", fmt.Errorf("Failed to prepare source files in temp template file store: '%w'.", err)
	}

	this.store[id] = tempDir
	return tempDir, nil
}

// Remove all the temp template files.
func (this *TemplateFileStore) Close() {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, path := range this.store {
		util.RemoveDirent(path)
	}

	this.store = make(map[string]string)
}

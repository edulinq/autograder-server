package disk

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/eriq-augustine/autograder/util"
)

const DISK_DB_TASKS_FILENAME = "tasks.json"

func (this *backend) LogTaskCompletion(courseID string, taskID string, instance time.Time) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	log, err := this.getTaskLog(courseID)
	if err != nil {
		return err
	}

	log[taskID] = instance

	err = this.writeTaskLog(courseID, log)
	if err != nil {
		return err
	}

	return nil
}

func (this *backend) GetLastTaskCompletion(courseID string, taskID string) (time.Time, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	log, err := this.getTaskLog(courseID)
	if err != nil {
		return time.Time{}, err
	}

	instance, exists := log[taskID]
	if !exists {
		return time.Time{}, nil
	}

	return instance, nil
}

func (this *backend) getTasksPathFromID(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_TASKS_FILENAME)
}

func (this *backend) getTaskLog(courseID string) (map[string]time.Time, error) {
	path := this.getTasksPathFromID(courseID)

	var log map[string]time.Time
	if util.PathExists(path) {
		err := util.JSONFromFile(path, &log)
		if err != nil {
			return nil, fmt.Errorf("Failed to read task log '%s': '%w'.", path, err)
		}
	} else {
		log = make(map[string]time.Time)
	}

	return log, nil
}

func (this *backend) writeTaskLog(courseID string, log map[string]time.Time) error {
	path := this.getTasksPathFromID(courseID)

	err := util.MkDir(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("Failed to create directory for task log '%s': '%w'.", path, err)
	}

	err = util.ToJSONFileIndent(log, path)
	if err != nil {
		return fmt.Errorf("Failed to write task log '%s': '%w'.", path, err)
	}

	return nil
}

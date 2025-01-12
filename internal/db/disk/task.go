package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_ACTIVE_TASKS_FILENAME = "active-tasks.json"

func (this *backend) GetActiveCourseTasks(course *model.Course) (map[string]*model.FullScheduledTask, error) {
	this.tasksLock.RLock()
	defer this.tasksLock.RUnlock()

	allTasks, err := this.getTasks()
	if err != nil {
		return nil, err
	}

	courseTasks := make(map[string]*model.FullScheduledTask, len(allTasks))
	for hash, task := range allTasks {
		if (task.Source == model.TaskSourceCourse) && (task.CourseID == course.ID) {
			courseTasks[hash] = task
		}
	}

	return courseTasks, nil
}

func (this *backend) GetNextActiveTask() (*model.FullScheduledTask, error) {
	this.tasksLock.RLock()
	defer this.tasksLock.RUnlock()

	allTasks, err := this.getTasks()
	if err != nil {
		return nil, err
	}

	var nextTask *model.FullScheduledTask = nil
	for _, task := range allTasks {
		if (nextTask == nil) || (nextTask.NextRunTime > task.NextRunTime) {
			nextTask = task
		}
	}

	return task, nil
}

func (this *backend) UpsertActiveTasks(upsertTasks map[string]*model.FullScheduledTask) error {
	this.tasksLock.Lock()
	defer this.tasksLock.Unlock()

	allTasks, err := this.getTasks()
	if err != nil {
		return nil, err
	}

	for hash, upsertTask := range upsertTasks {
		if upsertTask == nil {
			delete(allTasks, hash)
		} else {
			allTasks[hash] = upsertTask
		}
	}

	return writeTasks(allTasks)
}

func (this *backend) getActiveTasksPath() string {
	return filepath.Join(this.baseDir, DISK_DB_ACTIVE_TASKS_FILENAME)
}

func (this *backend) getTasks() (map[string]*model.FullScheduledTask, error) {
	task := make(map[string]*model.FullScheduledTask, 0)

	path := getActiveTasksPath()
	if !util.PathExists(path) {
		return tasks, nil
	}

	err := util.JSONFromFile(path, &tasks)
	if err != nil {
		return nil, fmt.Errorf("Failed to read active tasks file '%s': '%w'.", path, err)
	}

	return tasks, nil
}

func (this *backend) writeTasks(task map[string]*model.FullScheduledTask) error {
	path := getActiveTasksPath()

	err := util.MkDir(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("Failed to create directory for active tasks file '%s': '%w'.", path, err)
	}

	err = util.ToJSONFileIndent(tasks, path)
	if err != nil {
		return fmt.Errorf("Failed to write active tasks file '%s': '%w'.", path, err)
	}

	return nil
}

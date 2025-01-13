package tasks

import (
	"sync"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

const MIN_WAIT_MSECS = 1000

var (
	lock             sync.Mutex
	enableTaskEngine bool = false
)

func Start() {
	lock.Lock()
	defer lock.Unlock()

	enableTaskEngine = true
	go runTasks()
}

// Stop any future tasks from running.
// Once this returns, no more tasks will be run (and none will be in-progress).
// No tasks will be interrupted, so this call may block until any in-progress tasks stop.
func Stop() {
	lock.Lock()
	defer lock.Unlock()

	enableTaskEngine = false
}

// Run tasks until Stop() is called.
// Even when Stop is called, no tasks will be interrupted.
func runTasks() {
	for enableTaskEngine {
		runNextTask()

		task, err := db.GetNextActiveTask()
		if err != nil {
			log.Error("Failed to get next active task.", err)
		}

		waitTimeMSecs := int64(config.TASK_MAX_WAIT_SECS.Get() * 1000)
		if task != nil {
			taskWaitTimeMSecs := int64(max(MIN_WAIT_MSECS, task.NextRunTime.ToMSecs()-timestamp.Now().ToMSecs()))
			waitTimeMSecs = min(waitTimeMSecs, taskWaitTimeMSecs)
		}

		log.Trace("Waiting for next task check.", log.NewAttr("time-msecs", waitTimeMSecs))
		time.Sleep(time.Duration(waitTimeMSecs) * time.Millisecond)
	}
}

func runNextTask() {
	lock.Lock()
	defer lock.Unlock()

	if !enableTaskEngine {
		return
	}

	if config.NO_TASKS.Get() {
		return
	}

	task, err := db.GetNextActiveTask()
	if err != nil {
		log.Error("Failed to get next active task for running.", err)
		return
	}

	if (task == nil) || task.Disabled {
		return
	}

	now := timestamp.Now()
	if task.NextRunTime > now {
		return
	}

	log.Debug("Task started.", task)
	runTask(task)
	log.Debug("Task finished.", task)

	task.AdvanceRunTimes()

	err = db.UpsertActiveTask(task)
	if err != nil {
		log.Error("Failed to save task.", err)
		return
	}
}

func runTask(task *model.FullScheduledTask) {
	if task == nil {
		return
	}

	defer func() {
		value := recover()
		if value == nil {
			return
		}

		log.Error("Task paniced.", task, log.NewAttr("recover-value", value))
	}()

	var err error = nil

	switch task.Type {
	case model.TaskTypeCourseBackup:
		err = RunCourseBackupTask(task)
	case model.TaskTypeCourseEmailLogs:
		err = RunCourseEmailLogsTask(task)
	case model.TaskTypeCourseReport:
		err = RunCourseReportTask(task)
	case model.TaskTypeCourseScoringUpload:
		err = RunCourseScoringUploadTask(task)
	case model.TaskTypeCourseUpdate:
		err = RunCourseUpdateTask(task)
	case model.TaskTypeTest:
		err = RunTestTask(task)
	default:
		log.Error("Unknown task type.", task)
		return
	}

	if err != nil {
		log.Error("Failed to run task.", task, err)
	}
}

// Implementations for running tasks.
package task

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/maps"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/timestamp"
)

var timersLock sync.Mutex

type timerInfo struct {
	ID       string
	TaskID   string
	CourseID string
	Timer    *time.Timer
	Lock     *sync.Mutex
	// If this specific timer instance has been stopped, do not run the task.
	// This catches a different case than stoppedTasks
	// (which prevents a task from re-scheduling).
	Stopped bool
}

// {courseID: {ID: timerInfo, ...}, ...}
var timers map[string]map[string]*timerInfo = make(map[string]map[string]*timerInfo)

// Stopped tasks should not be rescheduled automatically.
// Manually scheduling a stopped task (via Schedule()) will remove it from this map.
// The key (timerID) is the same as timerInfo.ID.
var stoppedTasks map[string]bool = make(map[string]bool)

// The boolean return indicates if a task should be scheduled again.
type RunFunc func(*model.Course, tasks.ScheduledTask) (bool, error)

func init() {
	go watchHandle()
}

func Schedule(course *model.Course, target tasks.ScheduledTask) error {
	if target.IsDisabled() || config.NO_TASKS.Get() {
		return nil
	}

	var runFunc RunFunc
	switch target.(type) {
	case *tasks.BackupTask:
		runFunc = RunBackupTask
	case *tasks.CourseUpdateTask:
		runFunc = RunCourseUpdateTask
	case *tasks.EmailLogsTask:
		runFunc = RunEmailLogsTask
	case *tasks.ReportTask:
		runFunc = RunReportTask
	case *tasks.ScoringUploadTask:
		runFunc = RunScoringUploadTask
	case *tasks.TestTask:
		runFunc = RunTestTask
	default:
		return fmt.Errorf("Unknown task type: %t (%v).", target, target)
	}

	// Does this task need to be run right now
	// to cachup for lost time (e.g. the server being down).
	catchup, err := checkForCatchup(course.GetID(), target)
	if err != nil {
		return err
	}

	if catchup {
		// Special ID for catchup tasks.
		timerID := fmt.Sprintf("%s::catchup", target.GetID())

		err := scheduleTask(course.GetID(), target, timerID, runFunc, nil)
		if err != nil {
			return fmt.Errorf("Failed to schedule catchup task (%s): '%w'.", target.GetID(), err)
		}
	}

	for i, when := range target.GetTimes() {
		// ID unique to every (task, timer).
		timerID := fmt.Sprintf("%s::%03d", target.GetID(), i)

		timersLock.Lock()
		delete(stoppedTasks, timerID)
		timersLock.Unlock()

		err := scheduleTask(course.GetID(), target, timerID, runFunc, when)
		if err != nil {
			return fmt.Errorf("Failed to schedule task (%s): '%w'.", target.GetID(), err)
		}
	}

	return nil
}

// Check to see if it has been too long since this task has been run.
// Do this by getting the minimum duration for all the task's timers,
// and seeing if it has been at least that long since the task has been run.
// Return true of a catchup task needs to be run.
func checkForCatchup(courseID string, target tasks.ScheduledTask) (bool, error) {
	minDuration, exists := target.GetMinDurationMS()
	if !exists {
		return false, nil
	}

	lastRunTime, err := db.GetLastTaskCompletion(courseID, target.GetID())
	if err != nil {
		return false, fmt.Errorf("Failed to get last time task was run: '%w'.", err)
	}

	// Don't catchup if the task has never been run.
	if lastRunTime.IsZero() {
		return false, nil
	}

	currentDuration := timestamp.Now() - lastRunTime
	return (currentDuration.ToMSecs() > minDuration), nil
}

// Schedule a task.
// |when| will be nil on a catchup task.
func scheduleTask(courseID string, target tasks.ScheduledTask, timerID string, runFunc RunFunc, when *common.ScheduledTime) error {
	timersLock.Lock()
	defer timersLock.Unlock()

	// Ensure this task has not been stopped,
	// and therefore should not be scheduled.
	stopped, exists := stoppedTasks[timerID]
	if exists && stopped {
		return nil
	}

	// In addition to the general timers lock, create a lock just for this timer that prevents
	// a scheduled task from calling too soon (before the timerInfo is setup).
	taskLock := &sync.Mutex{}
	taskLock.Lock()
	defer taskLock.Unlock()

	nextRunInMS := int64(5)
	if when != nil {
		nextRunTime := when.ComputeNextTimeFromNow()
		nextRunInMS = nextRunTime.ToMSecs() - timestamp.Now().ToMSecs()
	}

	timer := time.AfterFunc(time.Duration(nextRunInMS*int64(time.Millisecond)), func() {
		// Ensure that this task does not start too quickly.
		// We will acquire this lock for the duration of the task run later.
		taskLock.Lock()
		taskLock.Unlock()

		reschedule := runTask(courseID, target, timerID, runFunc)

		if !reschedule {
			return
		}

		// Do not reschedule catchup tasks.
		if when == nil {
			return
		}

		// Schedule the next run.
		err := scheduleTask(courseID, target, timerID, runFunc, when)
		if err != nil {
			log.Error("Failed to reschedule task.", err,
				log.NewCourseAttr(courseID), log.NewAttr("task", target.GetID()), log.NewAttr("when", when.String()))
		}
	})

	_, ok := timers[courseID]
	if !ok {
		timers[courseID] = make(map[string]*timerInfo)
	}

	timers[courseID][timerID] = &timerInfo{
		ID:       timerID,
		TaskID:   target.GetID(),
		CourseID: courseID,
		Timer:    timer,
		Lock:     taskLock,
		Stopped:  false,
	}

	if when == nil {
		log.Trace("Catchup task scheduled.", log.NewCourseAttr(courseID), log.NewAttr("task", target.GetID()))
	} else {
		nextRunTime := timestamp.FromMSecs(timestamp.Now().ToMSecs() + nextRunInMS)
		log.Trace("Task scheduled.", log.NewCourseAttr(courseID), log.NewAttr("task", target.GetID()),
			log.NewAttr("timer-id", timerID), log.NewAttr("when", when.String()), log.NewAttr("next-time", nextRunTime))
	}

	return nil
}

func stopCoursesInternal(courseID string, acquireTimersLock bool, acquireTimerSpecificLock bool) {
	if acquireTimersLock {
		timersLock.Lock()
		defer timersLock.Unlock()
	}

	for _, timerInfo := range timers[courseID] {
		// Acquiring the lock ensure that we are not interrupting an active task.
		// If we get the lock before a scheduled task, then that task will never run.
		if acquireTimerSpecificLock {
			timerInfo.Lock.Lock()
			defer timerInfo.Lock.Unlock()
		}

		timerInfo.Stopped = true
		stoppedTasks[timerInfo.ID] = true
		timerInfo.Timer.Stop()

		log.Debug("Task stopped.", log.NewCourseAttr(courseID), log.NewAttr("task", timerInfo.TaskID), log.NewAttr("timer-id", timerInfo.ID))
	}

	delete(timers, courseID)
}

// Stop all the tasks associated with this course.
// This will block until all tasks have been stopped.
// Will wait for any already running tasks to finish.
func StopCourse(courseID string) {
	stopCoursesInternal(courseID, true, true)
}

// Stop all the tasks.
// This will block until all tasks have been stopped.
// Will wait for any already running tasks to finish.
func StopAll() {
	timersLock.Lock()
	defer timersLock.Unlock()

	keys := maps.Keys(timers)
	for _, key := range keys {
		stopCoursesInternal(key, false, true)
	}
}

// The boolean indicates if the task should be scheduled again.
func runTask(courseID string, target tasks.ScheduledTask, timerID string, runFunc RunFunc) bool {
	target.GetLock().Lock()
	defer target.GetLock().Unlock()

	taskID := target.GetID()
	now := timestamp.Now()

	info := getTimerInfo(courseID, timerID)
	if info == nil {
		return true
	}

	if info.Stopped {
		return true
	}

	info.Lock.Lock()
	defer info.Lock.Unlock()

	lastRunTime, err := db.GetLastTaskCompletion(courseID, taskID)
	if err != nil {
		log.Error("Failed to get last time task was run.", err, log.NewCourseAttr(courseID), log.NewAttr("task", taskID))
		// Keep trying to run the task.
	}

	// Skip this task if it was run too recently.
	lastRunDurationMS := int64(now.ToMSecs() - lastRunTime.ToMSecs())
	if lastRunDurationMS < int64(config.TASK_MIN_REST_SECS.Get()*1000) {
		log.Trace("Skipping task run, last run was too recent.",
			log.NewCourseAttr(courseID), log.NewAttr("task", taskID), log.NewAttr("last-run", lastRunTime))
		return true
	}

	log.Debug("Task started.", log.NewCourseAttr(courseID), log.NewAttr("task", taskID), log.NewAttr("timer", timerID))

	course, err := db.GetCourse(courseID)
	if err != nil {
		log.Error("Failed to get course for task.", err, log.NewCourseAttr(courseID), log.NewAttr("task", taskID))
		return true
	}

	if course == nil {
		log.Error("Could not find course for task.", log.NewCourseAttr(courseID), log.NewAttr("task", taskID))
		return true
	}

	reschedule, err := invokeRunFunc(course, target, runFunc)
	if err != nil {
		log.Error("Task run failed.", err, course, log.NewAttr("task", taskID))
		return true
	}

	log.Debug("Task finished.", course, log.NewAttr("task", taskID))

	err = db.LogTaskCompletion(courseID, taskID, now)
	if err != nil {
		log.Error("Failed to log task completion.", err, course, log.NewAttr("task", taskID))
		return reschedule
	}

	return reschedule
}

// Actually run the run func (and recover if necessary).
func invokeRunFunc(course *model.Course, target tasks.ScheduledTask, runFunc RunFunc) (reschedule bool, err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("Task paniced: '%v'.", value)
	}()

	reschedule, err = runFunc(course, target)
	return
}

// Should a task run?
func getTimerInfo(courseID string, timerID string) *timerInfo {
	timersLock.Lock()
	defer timersLock.Unlock()

	timers, ok := timers[courseID]
	if !ok {
		// No course timers are found, this course must have been stopped.
		return nil
	}

	info, ok := timers[timerID]
	if !ok {
		// This timer cannot be found, it must have been stopped.
		return nil
	}

	return info
}

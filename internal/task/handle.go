package task

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
)

// Forever watch the task handle created in model/tasks for requests.
// This handle allows for package to not rely on the task package explicitly for simple operations.
func watchHandle() {
	// We are now ready to accept calls from the model's handler.
	tasks.Handler.InitFromTask()

	for {
		select {
		case courseID := <-tasks.Handler.StopCourseChan:
			tasks.Handler.DoneChan <- handleStopCourse(courseID)
		case schedulePayload := <-tasks.Handler.ScheduleChan:
			tasks.Handler.DoneChan <- handleSchedule(schedulePayload)
		}
	}
}

func handleStopCourse(courseID string) error {
	StopCourse(courseID)
	return nil
}

func handleSchedule(schedulePayload *tasks.SchedulePayload) error {
	if schedulePayload == nil {
		return fmt.Errorf("Schedule payload is nil.")
	}

	course, ok := schedulePayload.Course.(*model.Course)
	if !ok {
		return fmt.Errorf("Schdeule payload course is not a *model.Course: %t (%v).", schedulePayload.Course, schedulePayload.Course)
	}

	return Schedule(course, schedulePayload.Target)
}

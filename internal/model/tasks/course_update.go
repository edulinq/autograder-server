package tasks

type CourseUpdateTask struct {
    *BaseTask
}

func (this *CourseUpdateTask) Validate(course TaskCourse) error {
    this.BaseTask.Name = "course-update";

    err := this.BaseTask.Validate(course);
    if (err != nil) {
        return err;
    }

    return nil;
}

package tasks

// A task for testing.
// This task will just call the provided function when run.
type TestTask struct {
	*BaseTask

	Func    func(any) error `json:"-"`
	Payload any             `json:"-"`
}

func (this *TestTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "test"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	return nil
}

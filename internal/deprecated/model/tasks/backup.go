package tasks

type BackupTask struct {
	*BaseTask

	Dest     string `json:"-"`
	BackupID string `json:"-"`
}

func (this *BackupTask) Validate(course TaskCourse) error {
	this.BaseTask.Name = "backup"

	err := this.BaseTask.Validate(course)
	if err != nil {
		return err
	}

	return nil
}

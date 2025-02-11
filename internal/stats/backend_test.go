package stats

type testBackend struct {
	system []*SystemMetrics
	course []*CourseMetric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) StoreCourseMetric(record *CourseMetric) error {
	this.course = append(this.course, record)
	return nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system: make([]*SystemMetrics, 0, 100),
		course: make([]*CourseMetric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

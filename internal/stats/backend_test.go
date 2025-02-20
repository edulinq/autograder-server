package stats

type testBackend struct {
	system  []*SystemMetrics
	course  []*CourseMetric
	request []*RequestMetric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) StoreCourseMetric(record *CourseMetric) error {
	this.course = append(this.course, record)
	return nil
}

func (this *testBackend) StoreRequestMetric(record *RequestMetric) error {
	this.request = append(this.request, record)
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

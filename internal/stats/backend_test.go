package stats

type testBackend struct {
	system  []*SystemMetrics
	course  []*CourseMetric
	request []*APIRequestMetric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) GetSystemStats(query Query) ([]*SystemMetrics, error) {
	return this.system, nil
}

func (this *testBackend) StoreCourseMetric(record *CourseMetric) error {
	this.course = append(this.course, record)
	return nil
}

func (this *testBackend) GetCourseMetrics(query CourseMetricQuery) ([]*CourseMetric, error) {
	return this.course, nil
}

func (this *testBackend) StoreAPIRequestMetric(record *APIRequestMetric) error {
	this.request = append(this.request, record)
	return nil
}

func (this *testBackend) GetAPIRequestMetrics(query Query) ([]*APIRequestMetric, error) {
	return this.request, nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system:  make([]*SystemMetrics, 0, 100),
		course:  make([]*CourseMetric, 0, 100),
		request: make([]*APIRequestMetric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

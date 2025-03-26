package stats

type testBackend struct {
	system     []*SystemMetrics
	course     []*BaseMetric
	apiRequest []*BaseMetric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) GetSystemStats(query Query) ([]*SystemMetrics, error) {
	return this.system, nil
}

func (this *testBackend) StoreCourseMetric(record *BaseMetric) error {
	this.course = append(this.course, record)
	return nil
}

func (this *testBackend) GetCourseMetrics(query MetricQuery) ([]*BaseMetric, error) {
	return this.course, nil
}

func (this *testBackend) StoreAPIRequestMetric(record *BaseMetric) error {
	this.apiRequest = append(this.apiRequest, record)
	return nil
}

func (this *testBackend) GetAPIRequestMetrics(query MetricQuery) ([]*BaseMetric, error) {
	return this.apiRequest, nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system:     make([]*SystemMetrics, 0, 100),
		course:     make([]*BaseMetric, 0, 100),
		apiRequest: make([]*BaseMetric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

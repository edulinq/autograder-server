package stats

type testBackend struct {
	system []*SystemMetrics
	course []*BaseMetric
	// apiRequest []*BaseMetric
	metric []*BaseMetric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) GetSystemStats(query Query) ([]*SystemMetrics, error) {
	return this.system, nil
}

func (this *testBackend) StoreMetric(record *BaseMetric) error {
	this.metric = append(this.metric, record)
	return nil
}

func (this *testBackend) GetMetrics(query MetricQuery) ([]*BaseMetric, error) {
	return this.metric, nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system: make([]*SystemMetrics, 0, 100),
		metric: make([]*BaseMetric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

package stats

type testBackend struct {
	system []*SystemMetrics
	metric []*Metric
}

func (this *testBackend) StoreSystemStats(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func (this *testBackend) GetSystemStats(query Query) ([]*SystemMetrics, error) {
	return this.system, nil
}

func (this *testBackend) StoreMetric(record *Metric) error {
	this.metric = append(this.metric, record)
	return nil
}

func (this *testBackend) GetMetrics(query Query) ([]*Metric, error) {
	return this.metric, nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system: make([]*SystemMetrics, 0, 100),
		metric: make([]*Metric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

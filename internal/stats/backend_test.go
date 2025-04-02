package stats

type testBackend struct {
	metrics []*Metric
}

func (this *testBackend) StoreMetric(record *Metric) error {
	this.metrics = append(this.metrics, record)
	return nil
}

func (this *testBackend) GetMetrics(query Query) ([]*Metric, error) {
	return this.metrics, nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		metrics: make([]*Metric, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

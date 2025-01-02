package stats

type testBackend struct {
	system []*SystemMetrics
}

func (this *testBackend) StoreSystemMetrics(record *SystemMetrics) error {
	this.system = append(this.system, record)
	return nil
}

func makeTestBackend() *testBackend {
	return &testBackend{
		system: make([]*SystemMetrics, 0, 100),
	}
}

func clearBackend() {
	backend = nil
}

package stats

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

type Metric interface {
	GetTime() timestamp.Timestamp
}

type BaseMetric struct {
	Time timestamp.Timestamp `json:"time"`
}

func (this BaseMetric) GetTime() timestamp.Timestamp {
	return this.Time
}

type SystemMetrics struct {
	BaseMetric

	CPUPercent       float64 `json:"cpu-percent"`
	MemPercent       float64 `json:"mem-percent"`
	NetBytesSent     uint64  `json:"net-bytes-sent"`
	NetBytesReceived uint64  `json:"net-bytes-received"`
}

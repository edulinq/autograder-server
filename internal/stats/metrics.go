package stats

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

type Metric interface {
	GetTimestamp() timestamp.Timestamp
}

type BaseMetric struct {
	Timestamp timestamp.Timestamp `json:"timestamp"`

	// Additional attributes that are not standard enough to be formalized in fields.
	Attributes map[string]any `json:"attributes,omitempty"`
}

type CourseAssignmentEmailMetric struct {
	CourseID     string `json:"course,omitempty"`
	AssignmentID string `json:"assignment,omitempty"`
	UserEmail    string `json:"user,omitempty"`
}

func (this BaseMetric) GetTimestamp() timestamp.Timestamp {
	return this.Timestamp
}

type SystemMetrics struct {
	BaseMetric

	CPUPercent       float64 `json:"cpu-percent"`
	MemPercent       float64 `json:"mem-percent"`
	NetBytesSent     uint64  `json:"net-bytes-sent"`
	NetBytesReceived uint64  `json:"net-bytes-received"`
}

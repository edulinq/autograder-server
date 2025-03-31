package stats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var (
	systemContextLock sync.Mutex
	ctx               context.Context    = nil
	cancelFunc        context.CancelFunc = nil
	cancelWait        *sync.WaitGroup    = nil

	getStatsLock      sync.Mutex
	lastBytesSent     uint64 = 0
	lastBytesReceived uint64 = 0
)

type SystemMetrics struct {
	Metric

	CPUPercent       float64 `json:"cpu-percent"`
	MemPercent       float64 `json:"mem-percent"`
	NetBytesSent     uint64  `json:"net-bytes-sent"`
	NetBytesReceived uint64  `json:"net-bytes-received"`
}

func collectSystemStats(systemIntervalMS int) {
	if backend == nil {
		log.Error("Stats backend is nil, cannot collect system stats.")
		return
	}

	// Signal that we have fully stopped collecting stats.
	cancelWait.Add(1)
	defer cancelWait.Done()

	for ctx != nil {
		// Check (and return) if the context was canceled.
		// Otherwise, gather system metrics again.
		// Note that the natural wait/sleep in
		// getSystemMetrics() of systemIntervalMS controls the timing of this loop.
		select {
		case <-ctx.Done():
			return
		default:
			if backend == nil {
				return
			}

			stats, err := getSystemMetrics(systemIntervalMS)
			if err != nil {
				log.Warn("Failed to collect system stats.", err)
				continue
			}

			// Don't store if the collection was already canceled.
			if ctx.Err() != nil {
				return
			}

			err = storeSystemStats(stats)
			if err != nil {
				log.Error("Failed to store system stats.", err)
				continue
			}
		}
	}
}

// Start collecting system stats.
func startSystemStatsCollection(systemIntervalMS int) {
	systemContextLock.Lock()
	defer systemContextLock.Unlock()

	if ctx != nil {
		// Already collecting stats.
		return
	}

	ctx, cancelFunc = context.WithCancel(context.Background())
	cancelWait = &sync.WaitGroup{}

	go collectSystemStats(systemIntervalMS)
}

func stopSystemStatsCollection() {
	// Note that we are not deferring an unlock.
	systemContextLock.Lock()

	if ctx == nil {
		// Already done collecting stats.
		systemContextLock.Unlock()
		return
	}

	// Cancel the collection.
	cancelFunc()

	// In the background, hold the lock until collection is complete.
	go func() {
		defer systemContextLock.Unlock()

		// Wait for completion.
		cancelWait.Wait()

		// Cleanup
		ctx = nil
		cancelFunc = nil
		cancelWait = nil
	}()
}

// Get the system metrics.
// Specific metrics have quirks the caller should know about.
//
// CPU:
// To get CPU statistics, we have to make an internal query, wait, make another internal query, and compare the results.
// The provided interval is the time (in MS) that should be waited between the two internal queries
// (longer waits generally yield more reliable results).
// This function will block for this entire process.
//
// Network:
// All known network interfaces will be summed.
// Additionally, the numbers provided will be the total number of bytes since the last call to this function,
// or zero if this is the first call.
func getSystemMetrics(intervalMS int) (*SystemMetrics, error) {
	getStatsLock.Lock()
	defer getStatsLock.Unlock()

	cpuMetrics, err := cpu.Percent(time.Millisecond*time.Duration(intervalMS), false)
	if err != nil {
		return nil, fmt.Errorf("Failed to get CPU metrics: '%w'.", err)
	}

	memMetrics, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("Failed to get memory metrics: '%w'.", err)
	}

	netMetrics, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("Failed to get net metrics: '%w'.", err)
	}

	var bytesSentDelta uint64 = 0
	if lastBytesSent != 0 {
		bytesSentDelta = netMetrics[0].BytesSent - lastBytesSent
	}

	var bytesReceivedDelta uint64 = 0
	if lastBytesReceived != 0 {
		bytesReceivedDelta = netMetrics[0].BytesRecv - lastBytesReceived
	}

	lastBytesSent = netMetrics[0].BytesSent
	lastBytesReceived = netMetrics[0].BytesRecv

	results := SystemMetrics{
		Metric: Metric{
			Timestamp: timestamp.Now(),
		},
		CPUPercent:       util.RoundWithPrecision(cpuMetrics[0], 2),
		MemPercent:       util.RoundWithPrecision(memMetrics.UsedPercent, 2),
		NetBytesSent:     bytesSentDelta,
		NetBytesReceived: bytesReceivedDelta,
	}

	return &results, nil
}

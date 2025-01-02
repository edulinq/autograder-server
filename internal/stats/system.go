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
	ctx        context.Context    = nil
	cancelFunc context.CancelFunc = nil
	cancelWait *sync.WaitGroup    = nil

	statsLock         sync.Mutex
	lastBytesSent     uint64 = 0
	lastBytesReceived uint64 = 0
)

func collectSystemStats(systemIntervalMS int) {
	if backend == nil {
		log.Error("Stats backend is nil, cannot collect system stats.")
		return
	}

	// Signal that we have fully stopped collecting stats.
	cancelWait.Add(1)
	defer cancelWait.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			stats, err := GetSystemMetrics(systemIntervalMS)
			if err != nil {
				log.Warn("Failed to collect system stats.", err)
				continue
			}

			err = backend.StoreSystemMetrics(stats)
			if err != nil {
				log.Error("Failed to store system stats.", err)
				continue
			}
		}
	}
}

// Start collecting system stats.
func startSystemStatsCollection(systemIntervalMS int) {
	statsLock.Lock()
	defer statsLock.Unlock()

	if ctx != nil {
		// Already collecting stats.
		return
	}

	ctx, cancelFunc = context.WithCancel(context.Background())
	cancelWait = &sync.WaitGroup{}

	go collectSystemStats(systemIntervalMS)
}

func stopSystemStatsCollection() {
	statsLock.Lock()
	defer statsLock.Unlock()

	if ctx == nil {
		// Already done collecting stats.
		return
	}

	// Cancel and wait for any in-progress collection to stop.
	cancelFunc()
	cancelWait.Wait()

	ctx = nil
	cancelFunc = nil
	cancelWait = nil
}

// Get the system metrics.
// Specific metrics have quirks the caller should know about.
//
// CPU:
// The get CPU statistics, we have to make an internal query, wait, make another internal query, and compare the results.
// The provided internal is the time (in MS) that should be waited between the two internal queries
// (longer waits generally yield more reliable results).
// This function will block for this entire process.
//
// Network:
// All known network interfaces will be averaged for the summed.
// Additionally, the numbers provided will be the total number of bytes since the last call to this function,
// or zero if this is the first call.
func GetSystemMetrics(intervalMS int) (*SystemMetrics, error) {
	statsLock.Lock()
	defer statsLock.Unlock()

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
		BaseMetric: BaseMetric{
			Time: timestamp.Now(),
		},
		CPUPercent:       util.RoundWithPrecision(cpuMetrics[0], 2),
		MemPercent:       util.RoundWithPrecision(memMetrics.UsedPercent, 2),
		NetBytesSent:     bytesSentDelta,
		NetBytesReceived: bytesReceivedDelta,
	}

	return &results, nil
}

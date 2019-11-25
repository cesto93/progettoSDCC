package monitor

import (
	"time"
)

type MetricData struct {
	Label string
	Timestamps []time.Time
	Values []float64 
}

type MonitorBridge interface {
	GetMetrics(startTime time.Time, endTime time.Time) ([]MetricData, error)
}
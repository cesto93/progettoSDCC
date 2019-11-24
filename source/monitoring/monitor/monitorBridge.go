package monitor

import (
	"time"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type MonitorBridge interface {
	GetMetrics(startTime time.Time, endTime time.Time) []*cloudwatch.MetricDataResult
}
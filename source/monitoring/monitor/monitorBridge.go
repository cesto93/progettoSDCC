package monitor

import (
	"time"
	"fmt"
)

type MetricData struct {
	Label string
	Timestamps []time.Time
	Values []interface{}
}

type MonitorBridge interface {
	GetMetrics(startTime time.Time, endTime time.Time) ([]MetricData, error)
}

func printMetricDatas(metricsDatas []MetricData){
    for i:=0; i<len(metricsDatas); i++{
        fmt.Println(metricsDatas[i].Label)
        for j:=0; j<len(metricsDatas[i].Timestamps); j++{
            fmt.Println("time: ", (metricsDatas[i].Timestamps[j]).String(),"value: ", metricsDatas[i].Values[j])
        }
    }
}
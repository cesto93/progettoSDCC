package appMetrics

import (
	"os"
	"fmt"
	"time"
	"progettoSDCC/source/utility"
	"progettoSDCC/source/monitoring/monitor"
)

func NewAppMetric(appName string, label string, value interface{}) monitor.MetricData {
	return monitor.MetricData  {
			Label : label,
			TagName : "App",
			TagValue : appName,
			Timestamps: []time.Time{time.Now()},
			Values : []interface{}{value},
		}
}

func AppendApplicationMetrics(path string, metrics []monitor.MetricData) error {
	data, err := ReadApplicationMetrics(path)
	last := metrics
	if err != nil {
		return fmt.Errorf("failed to append application metrics : %v\n", err)
	} 
	if data == nil {
		data = last
	} else {
		data = append(data, last...)
	}
	return utility.ExportJson(path, data)
}

func ReadApplicationMetrics(path string) ([]monitor.MetricData, error) {
	var res []monitor.MetricData
	err := utility.ImportJson(path, &res)
	if os.IsNotExist(err) {
		return nil, nil
	}
	os.Remove(path)
	return res, err
}
package appMetrics

import (
	"os"
	"fmt"
	"time"
	"progettoSDCC/source/utility"
	"progettoSDCC/source/monitoring/monitor"
)

func NewAppMetrics(appName string, labels []string, values []interface{}) []monitor.MetricData {
	var metrics []monitor.MetricData
	timestamp := []time.Time{time.Now()}
	for i,_ := range labels {
		metrics = append(metrics, monitor.MetricData  {
				Label : labels[i],
				TagName : "App",
				TagValue : appName,
				Timestamps: timestamp,
				Values : []interface{}{values[i]},
		})
	}
	return metrics
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
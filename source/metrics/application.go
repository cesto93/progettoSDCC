package metrics

import (
	"time"
	"progettoSDCC/source/utility"
)


type WordCountMetrics struct {
	WordElaboratedApplication int
	LatencyApplication time.Duration
	ThroughPutApplication int
	Workers int
}

type WordCountMetricsData struct {
	Metrics WordCountMetrics
	Timestamp time.Time
}

func AppendApplicationMetrics(path string, metrics WordCountMetrics) error {
	data, err := ReadApplicationMetrics(path)
	if err != nil {
		return err
	}
	data = append(data, WordCountMetricsData{metrics, time.Now()})
	return utility.ExportJson(path, data)
}

func ReadApplicationMetrics(path string) ([]WordCountMetricsData, error) {
	var res []WordCountMetricsData
	err := utility.ImportJson(path, &res)
	return res, err
}
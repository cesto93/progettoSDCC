package metrics

import (
	"time"
	"progettoSDCC/source/utility"
	"os"
)


type WordCountMetrics struct {
	WordElaboratedApplication int
	ElaborationTime float64 	//sec
	ThroughPutApplication float64 //(words/sec)
	Workers int
}

type WordCountMetricsData struct {
	Metrics WordCountMetrics
	Timestamp time.Time
}

func AppendApplicationMetrics(path string, metrics WordCountMetrics) error {
	data, err := ReadApplicationMetrics(path)
	last := WordCountMetricsData{metrics, time.Now()}
	if err != nil {
		if os.IsNotExist(err) {
			data = []WordCountMetricsData{last}
		} else {
			return err
		} 
	} else {
		data = append(data, last)
	}
	return utility.ExportJson(path, data)
}

func ReadApplicationMetrics(path string) ([]WordCountMetricsData, error) {
	var res []WordCountMetricsData
	err := utility.ImportJson(path, &res)
	return res, err
}
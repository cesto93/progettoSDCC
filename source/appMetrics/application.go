package appMetrics

import (
	"time"
	"os"
	"fmt"
	"progettoSDCC/source/utility"
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

		return fmt.Errorf("failed to append application metrics : %v\n", err)
	} 
	if data == nil {
		data = []WordCountMetricsData{last}	
	} else {
		data = append(data, last)
	}
	return utility.ExportJson(path, data)
}

func ReadApplicationMetrics(path string) ([]WordCountMetricsData, error) {
	var res []WordCountMetricsData
	err := utility.ImportJson(path, &res)
	if os.IsNotExist(err) {
		return nil, nil
	}
	return res, err
}
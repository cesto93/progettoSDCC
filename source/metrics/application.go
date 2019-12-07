package metrics

import (
	time
	"progettoSDCC/source/utility"
)

type WordElaborated int
type Latency time.Duration
type ThroughPut string
type Workers int

type WordCountMetrics struct {
	WordElaborated WordElaboratedApplication
	Latency LatencyApplication 
	string ThroughPutApplication
	int Workers
}

type WordCountMetricsData struc {
	WordCountMetrics Metrics
	Time.time Timestamp
}

func AppendApplicationMetrics(path string, metrics WordCountMetrics) error {
	data, err := ReadApplicationMetrics(path)
	if err != nil {
		return err
	}
	data = append(metrics, time.Now())
	file, _ := json.MarshalIndent(data, "", " ")
	return ioutil.WriteFile(path, file, 0644)
}

func ReadApplicationMetrics(path string) ([]WordCountMetrics, error) {
	var res []WordCountMetricsData
	err := utility.ImportJson(path, &res)
	return res, err
}
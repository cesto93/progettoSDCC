package monitor

import (
        "strconv"
	"io/ioutil"
        "fmt"
        "net/http"
        "time"
        "strings"
        "encoding/json"
        "progettoSDCC/source/utility"
)

type ResultMetric struct {
        Name string `json:"__name__"`
        Job string `json:"job"`
        Instance string `json:"instance"`
}

type Prometheus_Result struct {
        Metric ResultMetric `json:"metric"`
        Value [][]interface{} `json:"values"`
}

type Prometheus_Data struct {
        ResultType string `json:"resultType"`
        Result []Prometheus_Result `json:"result"`
}

type Prometheus_Resp struct {
        Status string `json:"status"`
        Data Prometheus_Data `json:"data"`
}

type Prometheus_Metric struct {
        Name string
        Url string
}

type PrometheusMonitor struct {
        Metric []Prometheus_Metric
}

//instant queries

/*func (monitor *PrometheusMonitor) getPrometheusMetricsInstant(PrometheusMetricsJsonPath string) ([]MetricData, error) {
        var metric []Prometheus_Metric
        var prom_resp Prometheus_Resp
        var result []MetricData
        var r MetricData

        for i:=0; i<len(monitor.Metric); i++{
                resp, err:= http.Get(monitor.Metric[i].Url)
                if err != nil {
                        fmt.Errorf("could not read time series value, %v ", err)
                }
                body, err := ioutil.ReadAll(resp.Body)
                err = json.Unmarshal(body, &prom_resp)
                if err != nil {
                        return nil, fmt.Errorf("could not read time series value, %v ", err)
                }
                for j:=0; j<len(prom_resp.Data.Result); j++{
                        //fmt.Println(prom_resp.Data.Result[j].Metric)
                        r.Values= make([]interface{}, 1)
                        r.Timestamps= make([]time.Time, 1)
                        r.Label = monitor.Metric[i].Name
                        r.Timestamps[0] = time.Unix(int64(prom_resp.Data.Result[j].Value[0].(float64)), 0)
                        r.Values[0] = prom_resp.Data.Result[j].Value[1]
                        resp.Body.Close()
                        result = append(result, r)
                }
        }
        printMetricDatas(result)
        return result, nil
}*/

//range queries

func (monitor *PrometheusMonitor) GetMetrics(startTime time.Time, endTime time.Time) ([]MetricData, error) {
        var prom_resp Prometheus_Resp
        var result []MetricData
        var r MetricData

        for i:=0; i<len(monitor.Metric); i++{
                req:=monitor.Metric[i].Url + "&start=" + startTime.Format(time.RFC3339) + "&end=" + endTime.Format(time.RFC3339) + "&step=1m"
                resp, err:= http.Get(req)
                if err != nil {
                        return nil, fmt.Errorf("could not get time series, %v ", err)
                }
                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        return nil, fmt.Errorf("could not read time series , %v ", err)
                }
                //fmt.Println(req)
                err = json.Unmarshal(body, &prom_resp)
                if err != nil {
                        return nil, fmt.Errorf("could not read time series value, %v ", err)
                }
                //fmt.Println(prom_resp)
                for k:=0; k<len(prom_resp.Data.Result); k++{
                        //fmt.Println(prom_resp.Data.Result[k].Metric.Instance)
                        r.Label = monitor.Metric[i].Name
                        r.TagName = "instance_ip"
                        r.TagValue = strings.Split(prom_resp.Data.Result[k].Metric.Instance, ":")[0]
                        r.Timestamps = make([]time.Time, len(prom_resp.Data.Result[k].Value))
                        r.Values = make([]interface{}, len(prom_resp.Data.Result[k].Value))
                        for j:=0; j<len(prom_resp.Data.Result[k].Value); j++{
                                r.Timestamps[j] = time.Unix(int64(prom_resp.Data.Result[k].Value[j][0].(float64)), 0)
                                r.Values[j], _ = strconv.ParseFloat(prom_resp.Data.Result[k].Value[j][1].(string), 64)
                        }
                        resp.Body.Close()
                        result = append(result, r)
                }
        }
        printMetricDatas(result)
        return result, nil
}

func NewPrometheus(PrometheusMetricsJsonPath string) *PrometheusMonitor {
        var metrics []Prometheus_Metric

        utility.CheckError(utility.ImportJson(PrometheusMetricsJsonPath, &metrics))
        return &PrometheusMonitor{metrics}
}
package monitor

import (
	"io/ioutil"
        "fmt"
        "net/http"
        "time"
        //"os"
        //"log"
        "encoding/json"
        "progettoSDCC/source/utility"
)

type Prometheus_Result struct {
        Metric interface{} `json:"metric"`
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

//instant queries

/*func GetPrometheusMetrics(PrometheusMetricsJsonPath string) ([]MetricData, error) {
        var metric []Prometheus_Metric
        var prom_resp Prometheus_Resp
        var result []MetricData
        var r MetricData
        var err error = nil

        utility.ImportJson(PrometheusMetricsJsonPath, &metric)
        for i:=0; i<len(metric); i++{
                resp, err:= http.Get(metric[i].Url)
                if err != nil {
                        fmt.Errorf("could not read time series value, %v ", err)
                }
                body, err := ioutil.ReadAll(resp.Body)
                err = json.Unmarshal(body, &prom_resp)
                if err != nil {
                        fmt.Errorf("could not read time series value, %v ", err)
                }
                for j:=0; j<len(prom_resp.Data.Result); j++{
                        //fmt.Println(prom_resp.Data.Result[j].Metric)
                        r.Values= make([]interface{}, 1)
                        r.Timestamps= make([]time.Time, 1)
                        r.Label = metric[i].Name
                        r.Timestamps[0] = time.Unix(int64(prom_resp.Data.Result[j].Value[0].(float64)), 0)
                        r.Values[0] = prom_resp.Data.Result[j].Value[1]
                        resp.Body.Close()
                        result = append(result, r)
                }
        }
        printMetricDatas(result)
        return result, err
}*/

func GetPrometheusMetricsRange(PrometheusMetricsJsonPath string, startTime time.Time, endTime time.Time) ([]MetricData, error) {
        var metric []Prometheus_Metric
        var prom_resp Prometheus_Resp
        var result []MetricData
        var r MetricData
        var err error = nil

        utility.ImportJson(PrometheusMetricsJsonPath, &metric)
        for i:=0; i<len(metric); i++{
                req:=metric[i].Url + "&start=" + startTime.Format(time.RFC3339) + "&end=" + endTime.Format(time.RFC3339) + "&step=1m"
                resp, err:= http.Get(req)
                if err != nil {
                        fmt.Errorf("could not read time series value, %v ", err)
                }
                body, err := ioutil.ReadAll(resp.Body)
                //fmt.Println(req)
                err = json.Unmarshal(body, &prom_resp)
                if err != nil {
                        fmt.Errorf("could not read time series value, %v ", err)
                }
                //fmt.Println(prom_resp)
                for k:=0; k<len(prom_resp.Data.Result); k++{
                        //fmt.Println(prom_resp.Data.Result[k])
                        r.Label = metric[i].Name
                        r.Timestamps = make([]time.Time, len(prom_resp.Data.Result[k].Value))
                        r.Values = make([]interface{}, len(prom_resp.Data.Result[k].Value))
                        for j:=0; j<len(prom_resp.Data.Result[k].Value); j++{
                                r.Timestamps[j] = time.Unix(int64(prom_resp.Data.Result[k].Value[j][0].(float64)), 0)
                                r.Values[j] = prom_resp.Data.Result[k].Value[j][1]
                        }
                        resp.Body.Close()
                        result = append(result, r)
                }
        }
        printMetricDatas(result)
        return result, err
}
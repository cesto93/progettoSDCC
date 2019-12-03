package monitor

import (
		//"io/ioutil"
        "context"
        "fmt"
        //"io"
        "time"
        //"os"
        //"log"
        //"encoding/json"
        "progettoSDCC/source/utility"


        monitoring "cloud.google.com/go/monitoring/apiv3"
        "github.com/golang/protobuf/ptypes/timestamp"
        "google.golang.org/api/iterator"
        monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
    ProjectID = "concise-faculty-246814"
    //GcloudMetricsJsonPath = "../../../configuration/metrics_gce.json"
    //InstancesJsonPath = "../../../configuration/instances_ids.json"
)

type Metric_Label struct {
    Name string
    Value string
}

type Metric struct {
    Name string
    Url string
    Labels []Metric_Label `json:"Labels"`
}

type gceMonitor struct{
    client *monitoring.MetricClient
    ctx context.Context
    metrics []Metric
    instances []string
}

// readTimeSeriesValue reads the TimeSeries for the value specified by metric type in a time window from the last 20 minutes.
func (monitor *gceMonitor) GetMetrics(startTime time.Time, endTime time.Time) ([]MetricData, error) {
        var req *monitoringpb.ListTimeSeriesRequest
        var result []MetricData
        var r MetricData
        var err error = nil

        for i:=0; i<len(monitor.metrics); i++ {
            metricType:=concatenate(monitor.metrics[i], monitor.instances)
            //fmt.Println("Metric: ", monitor.metrics[i].Name, ", Labels: ", monitor.metrics[i].Labels)

            req = &monitoringpb.ListTimeSeriesRequest{
                Name:   "projects/" + ProjectID,
                Filter: metricType,
                Interval: &monitoringpb.TimeInterval{
                        StartTime: &timestamp.Timestamp{Seconds: startTime.Unix()},
                        EndTime:   &timestamp.Timestamp{Seconds: endTime.Unix()},
                },
                //PageSize: 2,
            }
            iter := monitor.client.ListTimeSeries(monitor.ctx, req)
            for {
                resp, err := iter.Next()
                if err == iterator.Done {
                    break
                }
                if err != nil {
                    fmt.Errorf("could not read time series value, %v ", err)
                }
                r.Label= monitor.metrics[i].Name
                //r.Values= make([]interface{}, len(resp.GetPoints()))
                r.Timestamps= make([]time.Time, len(resp.GetPoints()))
                for k:=0; k<len(resp.GetPoints()); k++{
                    r.Timestamps[k]= time.Unix(resp.GetPoints()[k].GetInterval().GetEndTime().GetSeconds(), 0)
                    //r.Values[k]= resp.GetPoints()[k].GetValue()
                    r.Values =ParseValue(resp)
                }
                result= append(result, r)
            }
        }
        printMetricDatas(result)
        return result, err
}

func ParseValue(val *monitoringpb.TimeSeries) ([]interface{}){
    var v []interface{}
    v = make([]interface{}, len(val.GetPoints()))
    //fmt.Println(val.GetValueType())
    switch(string(val.GetValueType().String())){
    case "DOUBLE":
        for i:=0; i<len(val.GetPoints()); i++{
            v[i]=val.GetPoints()[i].GetValue().GetDoubleValue()
        }
    case "BOOL":
        for i:=0; i<len(val.GetPoints()); i++{
            v[i]=val.GetPoints()[i].GetValue().GetBoolValue()
        }
    case "INT64":
        for i:=0; i<len(val.GetPoints()); i++{
            v[i]=val.GetPoints()[i].GetValue().GetInt64Value()
        }
    case "STRING":
        for i:=0; i<len(val.GetPoints()); i++{
            v[i]=val.GetPoints()[i].GetValue().GetStringValue()
        }
    default:
        return nil
    }
    return v
}

func concatenate(metric Metric, instances []string) string {
	s:="metric.type=\"" + metric.Url + "\""
	for i:=0; i<len(metric.Labels); i++ {
		s+=" AND metric.labels." + metric.Labels[i].Name + "=\"" + metric.Labels[i].Value + "\""
	}
    s+=" AND resource.labels.instance_id= one_of("
    for i:=0; i<len(instances)-1; i++ {
        s+="\"" + instances[i] + "\" , "
    }
    s+="\"" + instances[len(instances)-1] + "\")"
	return s
}

func printInstancesIdsAndMetrics(instances []string, metric []Metric){
    for i:=0; i<len(instances); i++ {
        fmt.Println(instances[i])
    }
    for i:=0; i<len(metric); i++ {
        fmt.Println("Name: ", metric[i].Name)
        fmt.Println("Url: ", metric[i].Url)
        for k:=0; k<len(metric[i].Labels); k++ {
            fmt.Println(metric[i].Labels[k].Name, ": ", metric[i].Labels[k].Value)
        }
    }
}

func NewGce(GcloudMetricsJsonPath string, InstancesJsonPath string) *gceMonitor{
    var metrics []Metric
    var instances []string

    utility.ImportJson(GcloudMetricsJsonPath, &metrics)
    utility.ImportJson(InstancesJsonPath, &instances)

    //printInstancesIdsAndMetrics(instances, metricMonitors)

    ctx := context.Background()
    c, err := monitoring.NewMetricClient(ctx)
    if err != nil {
            fmt.Errorf("could not connect to google monitor")
    }
    return &gceMonitor{c, ctx, metrics, instances}
}

func (gcemonitor *gceMonitor) Test() {
    //gcemonitor:=NewGce()
    startTime := time.Now().UTC().Add(time.Minute * -5)
    endTime := time.Now().UTC()
    result, err:=gcemonitor.GetMetrics(startTime, endTime)
	if err != nil {
	   fmt.Errorf("could not read time series value, %v ", err)
    }
    printMetricDatas(result)
}
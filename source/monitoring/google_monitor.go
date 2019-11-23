package main

import (
		"io/ioutil"
        "context"
        "fmt"
        "io"
        "time"
        "net/http"
        //"os"
        //"log"
        "encoding/json"

        monitoring "cloud.google.com/go/monitoring/apiv3"
        "github.com/golang/protobuf/ptypes/timestamp"
        "google.golang.org/api/iterator"
        monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// readTimeSeriesFields reads the last 20 minutes of the given metric, aligns
// everything on 10 minute intervals, and combines values from different
// instances.
func readTimeSeriesFields(w io.Writer, projectID string) error {
        ctx := context.Background()
        client, err := monitoring.NewMetricClient(ctx)
        if err != nil {
                return fmt.Errorf("NewMetricClient: %v", err)
        }
        startTime := time.Now().UTC().Add(time.Minute * -20)
        endTime := time.Now().UTC()
        req := &monitoringpb.ListTimeSeriesRequest{
                Name:   "projects/" + projectID,
                Filter: `metric.type="compute.googleapis.com/instance/cpu/utilization"`,
                Interval: &monitoringpb.TimeInterval{
                        StartTime: &timestamp.Timestamp{
                                Seconds: startTime.Unix(),
                        },
                        EndTime: &timestamp.Timestamp{
                                Seconds: endTime.Unix(),
                        },
                },
                View: monitoringpb.ListTimeSeriesRequest_HEADERS,
        }
        fmt.Fprintln(w, "Found data points for the following instances:")
        it := client.ListTimeSeries(ctx, req)
        for {
                resp, err := it.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return fmt.Errorf("could not read time series value: %v", err)
                }
                fmt.Fprintf(w, "\t%v\n", resp.GetMetric().GetLabels()["instance_name"])
        }
        fmt.Fprintln(w, "Done")
        return nil
}

// readTimeSeriesValue reads the TimeSeries for the value specified by metric type in a time window from the last 20 minutes.
func readTimeSeriesValue(projectID string, metricType string) (*monitoring.TimeSeriesIterator, error) {
        ctx := context.Background()
        c, err := monitoring.NewMetricClient(ctx)
        if err != nil {
                return nil, err
        }
        startTime := time.Now().UTC().Add(time.Minute * -5).Unix()
        endTime := time.Now().UTC().Unix()

        req := &monitoringpb.ListTimeSeriesRequest{
                Name:   "projects/" + projectID,
                Filter: metricType,
                Interval: &monitoringpb.TimeInterval{
                        StartTime: &timestamp.Timestamp{Seconds: startTime},
                        EndTime:   &timestamp.Timestamp{Seconds: endTime},
                },
                //PageSize: 2,
        }
        iter := c.ListTimeSeries(ctx, req)

        /*for {
                resp, err := iter.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return nil, fmt.Errorf("could not read time series value, %v ", err)
                }
                for i:=0; i<len(resp.GetPoints()); i++{
                	fmt.Printf("%d %+v\n", i, resp.GetPoints()[i])
            	}
        }*/

        return iter, nil
}

type Metric_Label struct {
	Name string
	Value string
}

type Metric struct {
	Monitor string
	Name string
	Url string
	Labels []Metric_Label `json:"Labels"`
}

func concatenate(metric Metric) string {
	s:="metric.type=\"" + metric.Url + "\""
	for i:=0; i<len(metric.Labels); i++ {
		s+=" AND metric.labels." + metric.Labels[i].Name + "=\"" + metric.Labels[i].Value + "\""
	}
	return s
}

type Prometheus_Result struct {
	Metric interface{} `json:"metric"`
	Value []interface{} `json:"value"`
}

type Prometheus_Data struct {
	ResultType string `json:"resultType"`
	Result []Prometheus_Result `json:"result"`
}

type Prometheus_Resp struct {
	Status string `json:"status"`
	Data Prometheus_Data `json:"data"`
}

func main(){
	//readTimeSeriesFields(os.Stdout, "concise-faculty-246814")
	var metrics []Metric
	file, err := ioutil.ReadFile("metrics.json")
	if err != nil {
		fmt.Println("error:", err)
	}
	err = json.Unmarshal([]byte(file), &metrics)
	if err != nil {
		fmt.Println("error:", err)
	}
	for i:=0; i<len(metrics); i++ {
		fmt.Println("Monitor: ", metrics[i].Monitor)
		fmt.Println("Name: ", metrics[i].Name)
		fmt.Println("Url: ", metrics[i].Url)
		for j:=0; j<len(metrics[i].Labels); j++ {
			fmt.Println(metrics[i].Labels[j].Name, ": ", metrics[i].Labels[j].Value)
		}
	}

	for i:=0; i<len(metrics); i++{
		switch metrics[i].Monitor {
		case "Google":
			complete_metric:=concatenate(metrics[i])
			fmt.Println("Metric: ", metrics[i].Name, ", Labels: ", metrics[i].Labels)
			serie, err:=readTimeSeriesValue("concise-faculty-246814", complete_metric)
			if err != nil {
	        	fmt.Errorf("could not read time series value, %v ", err)
        	}
        	for {
	        	resp, err := serie.Next()
        		if err == iterator.Done {
	        		break
        		}
        		if err != nil {
	        		fmt.Errorf("could not read time series value, %v ", err)
        		}
        		//fmt.Printf("%+v\n", resp)
        		for i:=0; i<len(resp.GetPoints()); i++{
	                fmt.Printf("%d time: %+v, %v\n", i,  time.Unix(resp.GetPoints()[i].GetInterval().GetEndTime().GetSeconds(), 0), resp.GetPoints()[i].GetValue())
            	}
    		}
    	case "Prometheus":
    		var prom_resp Prometheus_Resp
    		resp, err:= http.Get(metrics[i].Url)
    		if err != nil {
	        		fmt.Errorf("could not read time series value, %v ", err)
        		}
        	defer resp.Body.Close()
        	body, err := ioutil.ReadAll(resp.Body)
        	err = json.Unmarshal(body, &prom_resp)
        	if err != nil {
	        		fmt.Errorf("could not read time series value, %v ", err)
        		}
        	fmt.Println("Metric: ", metrics[i].Name)
        	for j:=0; j<len(prom_resp.Data.Result); j++ {
    			fmt.Printf("%+v\n", prom_resp.Data.Result[j].Value[1])
    		}
    	default:
    		fmt.Printf("Invalid monitor: %s\n", metrics[i].Monitor)
    	}
	}
}
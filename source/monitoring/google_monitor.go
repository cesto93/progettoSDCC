package main

import (
        "context"
        "fmt"
        "io"
        "time"
        "os"
        "log"

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
func readTimeSeriesValue(projectID, metricType string) error {
        ctx := context.Background()
        c, err := monitoring.NewMetricClient(ctx)
        if err != nil {
                return err
        }
        startTime := time.Now().UTC().Add(time.Minute * -20).Unix()
        endTime := time.Now().UTC().Unix()

        req := &monitoringpb.ListTimeSeriesRequest{
                Name:   "projects/" + projectID,
                Filter: fmt.Sprintf("metric.type=\"%s\"", metricType),
                Interval: &monitoringpb.TimeInterval{
                        StartTime: &timestamp.Timestamp{Seconds: startTime},
                        EndTime:   &timestamp.Timestamp{Seconds: endTime},
                },
        }
        iter := c.ListTimeSeries(ctx, req)

        for {
                resp, err := iter.Next()
                if err == iterator.Done {
                        break
                }
                if err != nil {
                        return fmt.Errorf("could not read time series value, %v ", err)
                }
                log.Printf("%+v\n", resp)
        }

        return nil
}


func main(){
	readTimeSeriesFields(os.Stdout, "concise-faculty-246814")
	readTimeSeriesValue("concise-faculty-246814", "compute.googleapis.com/instance/cpu/utilization")
}

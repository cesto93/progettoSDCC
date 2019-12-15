package db

import(
	"fmt"
	"time"
	"encoding/json"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"progettoSDCC/source/monitoring/monitor"
)

type DbBridge struct {
	Addr string
	DB string
}

func NewDb(addr string, db string) *DbBridge {
	return &DbBridge{addr, db}
}

func (d *DbBridge) SaveMetrics(data []monitor.MetricData) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://" + d.Addr,
	})
	if err != nil {
		return fmt.Errorf("error in client generation %v", err)
	}
	defer c.Close()

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  d.DB,
		Precision: "s",
	})

	for _, metric := range data {
		for i, _ := range metric.Values {
			tags := map[string]string{metric.TagName: metric.TagValue}
			fields := map[string]interface{}{"value": metric.Values[i]}
			pt, err := client.NewPoint(metric.Label, tags, fields, metric.Timestamps[i])
			if err != nil {
				return fmt.Errorf("error in newpoint generation %v", err)
			}
			bp.AddPoint(pt)
		}
	}

	err = c.Write(bp)
	if err != nil {
		return fmt.Errorf("error in db write %v", err)
	}
	return nil
}

func (d *DbBridge) GetLastTimestamp(metricName string) (*time.Time, error){
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://" + d.Addr,
	})
	if err != nil {
		return nil, fmt.Errorf("error in client generation %v", err)
	}
	defer c.Close()

	q := client.NewQuery("SELECT last(value) FROM " + metricName, d.DB, "s")
	response, err := c.Query(q)
	if err != nil {
		return nil, fmt.Errorf("error in query %v", err)
	}
	if response.Error() != nil {
		return nil, fmt.Errorf("error in query %v", response.Error())
	}
	fmt.Println(response.Results[0])
	fmt.Println(response.Results[0].Series[0])
	json := response.Results[0].Series[0].Values[0][0].(json.Number)
	num,_ := json.Int64()
	last := time.Unix(num, 0)
	return &last, nil
}

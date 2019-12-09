package db

import(
	"fmt"
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
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://" + d.Addr,
	})
	if err != nil {
		return fmt.Errorf("error in client generation %v", err)
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  d.DB,
		Precision: "s",
	})

	// Create a point and add to batch
	/*tags := map[string]string{"cpu": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}*/
	for _, metric := range data {
		for i, _ := range metric.Values {
			fields := map[string]interface{}{"avg": metric.Values[i]}
			pt, err := client.NewPoint(metric.Label, nil, fields, metric.Timestamps[i])
			if err != nil {
				return fmt.Errorf("error in newpoint generation %v", err)
			}
			bp.AddPoint(pt)
		}
	}

	// Write the batch
	c.Write(bp)
	return nil
}

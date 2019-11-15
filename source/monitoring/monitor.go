package main

import (
 	"time"
 	"progettoSDCC/source/monitoring/aws_monitor"
 )

func main() {
 	/*startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")*/

 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:00:00+00:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T00:10:00+00:00")
 	
 	monitor := aws_monitor.New()

 	//fmt.Println(monitor.Metrics)
 	
	ec2Data := monitor.GetMetrics(startTime, endTime)
	aws_monitor.PrintMetrics(ec2Data)
	/*s3Data := cloudwatchGetMetrics(svc, s3Metrics, startTime, endTime, stat)
	cloudwatchPrintMetrics(s3Data)*/	
 }
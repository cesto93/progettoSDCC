 package main

 import (
 	"fmt"
 	"os"
 	"time"
 	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/cloudwatch"
 )

 const (
	awsRegion = "us-east-1"
)

func cloudwatchEC2Metric(metricName string, instanceIds []string, metricId string, stat string, startTime time.Time, endTime time.Time, 
							period int64) []*cloudwatch.MetricDataResult {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)

	namespace := "AWS/EC2"
	metricDimName := "InstanceId"

	dimensions := make([]*cloudwatch.Dimension, len(instanceIds))
	for i := 0; i < len(instanceIds); i++ {
		dimensions[i] = &cloudwatch.Dimension{
			Name : &metricDimName,
			Value : &instanceIds[i],
		}
	}

	query := &cloudwatch.MetricDataQuery{
		Id: &metricId,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &namespace,
				MetricName: &metricName,
				Dimensions: dimensions,
			},
			Period: &period,
			Stat:   &stat,
		},
	}

	resp, err := svc.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: []*cloudwatch.MetricDataQuery{query},
	})

	if err != nil {
		fmt.Println("Got error getting metric data")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return resp.MetricDataResults;
}

func cloudwatchEC2Metrics(instanceIds []string, stat string, startTime time.Time, endTime time.Time, period int64) {
	metricNames := []string{"CPUCreditUsage", "CPUCreditBalance", "CPUSurplusCreditBalance", "CPUSurplusCreditsCharged", 
	"NetworkPacketsIn", "NetworkPacketsOut", "CPUUtilization", "NetworkIn", "NetworkOut", "DiskReadBytes", "DiskWriteBytes", 
	"DiskReadOps", "DiskWriteOps", "StatusCheckFailed_System", "StatusCheckFailed_Instance", "StatusCheckFailed"}

	metricIds := []string{"cpucreditusage", "cpucreditbalance", "cpusurpluscreditbalance", "cpusurpluscreditscharged", 
	"netpackin", "netpackout", "cpuutil", "networkin", "networkout", "diskreadbytes", "diskwritebytes", 
	"diskreadops", "diskwriteops", "statuscheckfailed_system", "statuscheckfailed_instance", "statuscheckfailed"}

	for i, _ := range metricNames {
		results := cloudwatchEC2Metric(metricNames[i], instanceIds, metricIds[i], "Average", startTime, endTime, 300)
		for _, metricdata := range results {
		fmt.Println(*metricdata.Id)
		for j, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (*metricdata.Timestamps[j]).String(), *metricdata.Values[j])
			}
		}
	} 
}

 func main() {
 	startTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:30:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:40:00+02:00")
 	instanceIds := []string{"i-0706dcb2c513b981c"}

	cloudwatchEC2Metrics(instanceIds, "Average", startTime, endTime, 60)	
 }
 package main

 import (
 	"fmt"
 	"os"
 	"time"
 	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/cloudwatch"
 	"progettoSDCC/source/utility"
 )

 const (
 	awsRegion = "us-east-1"
 	metricJsonPath = "./metrics_aws.json"
 	EC2InstPath = "./EC2_inst.json"
 )

 type AWSMetric struct {
 	Name string
 	Id string
 	Monitoring bool 
 }

 type EC2Inst []string

func cloudwatchClient(awsRegion string) *cloudwatch.CloudWatch{
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)
 	return svc
}

func removeDisabledMetrics(metrics []AWSMetric) ([]string, []string) {
	names := make([]string, 0)
	ids := make([]string, 0)
	for _, metric := range metrics {
		if metric.Monitoring == true {
			names = append(names, metric.Name)
			ids = append(ids, metric.Id)
		}
	}
	return names, ids
}

func cloudwatchEC2Metrics(svc *cloudwatch.CloudWatch, instanceIds []string, metricNames []string, metricIds []string, 
							stat string, startTime time.Time, endTime time.Time, 
							period int64) []*cloudwatch.MetricDataResult {
	const namespace = "AWS/EC2"
	const metricDimName = "InstanceId"

	dimensions := make([]*cloudwatch.Dimension, len(instanceIds))
	for i := 0; i < len(instanceIds); i++ {
		dimensions[i] = &cloudwatch.Dimension{
			Name : aws.String(metricDimName),
			Value : &instanceIds[i],
		}
	}

	query := make([]*cloudwatch.MetricDataQuery, len(metricNames))
	for i := 0; i < len(metricNames); i++ {
		query[i] = &cloudwatch.MetricDataQuery{
			Id: &metricIds[i],
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  aws.String(namespace),
					MetricName: &metricNames[i],
					Dimensions: dimensions,
				},
				Period: &period,
				Stat:   &stat,
			},
		}
	}

	resp, err := svc.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: query,
	})

	if err != nil {
		fmt.Println("Got error getting metric data")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return resp.MetricDataResults;
}

func cloudwatchEC2PrintMetrics(instanceIds []string, stat string, startTime time.Time, endTime time.Time, period int64) {
	var metrics []AWSMetric

	utility.ImportJson(metricJsonPath, &metrics)
	metricNames, metricIds := removeDisabledMetrics(metrics)

	svc := cloudwatchClient(awsRegion)
	results := cloudwatchEC2Metrics(svc, instanceIds, metricNames, metricIds, "Average", startTime, endTime, 300)
	for _, metricdata := range results {
	fmt.Println(*metricdata.Label)
	for j, _ := range metricdata.Timestamps {
		fmt.Printf("%v %v\n", (*metricdata.Timestamps[j]).String(), *metricdata.Values[j])
		}
	} 
}

func cloudwatchS3Metrics(stat string, startTime time.Time, endTime time.Time, period int64) {

}

 func main() {
 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")
 	var instanceIds []string
 	utility.ImportJson(EC2InstPath, &instanceIds)

	cloudwatchEC2PrintMetrics(instanceIds, "Average", startTime, endTime, 60)	
 }
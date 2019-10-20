 package main

 import (
 	"context"
 	"fmt"
 	"os"
 	"time"

 	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/awserr"
 	"github.com/aws/aws-sdk-go/aws/request"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/s3"
 	"github.com/aws/aws-sdk-go/service/cloudwatch"
 )

 const (
	awsRegion = "us-east-1"
)

 //   # Upload myfile.txt to myBucket/myKey. Must complete within 10 minutes or will fail
func example_s3(bucket string, key string, timeout time.Duration) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := s3.New(sess)
 	ctx := context.Background()
 	var cancelFn func()
 	if timeout > 0 {
 		ctx, cancelFn = context.WithTimeout(ctx, timeout)
 	}
 	defer cancelFn() // Ensure the context is canceled to prevent leaking.
 	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
 		Bucket: aws.String(bucket),
 		Key:    aws.String(key),
 		Body:   os.Stdin,
 	})
 	if err != nil {
 		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
 			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
 		} else {
 			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
 		}
 		os.Exit(1)
 	}
}

func cloudwatchEC2Metric(metricName string, instanceIds []string, metricId string, stat string, startTime time.Time, endTime time.Time, 
							period int64) []*cloudwatch.MetricDataResult {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)

 	//endTime := time.Now()
	//duration, _ := time.ParseDuration("-5m")
	//startTime := endTime.Add(duration)
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

	for _, metricName := range metricNames {
		results := cloudwatchEC2Metric(metricName, instanceIds, metricName, "Average", startTime, endTime, 300)
		for _, metricdata := range results {
		fmt.Println(*metricdata.Id)
		for index, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (*metricdata.Timestamps[index]).String(), *metricdata.Values[index])
			}
		}
	} 
}

 func main() {
 	startTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:30:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:40:00+02:00")
 	instanceIds := []string{"i-0706dcb2c513b981c"}

 	/*results := cloudwatchEC2Metric("CPUUtilization", instanceIds, "cpu1", "Average", startTime, endTime, 300)
 	for _, metricdata := range results {
		fmt.Println(*metricdata.Id)
		for index, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (*metricdata.Timestamps[index]).String(), *metricdata.Values[index])
		}
	}*/

	cloudwatchEC2Metrics(instanceIds, "Average", startTime, endTime, 300)	
 }
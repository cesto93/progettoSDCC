 package main

 import (
 	"context"
 	"flag"
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

 // Usage:
 //   # Upload myfile.txt to myBucket/myKey. Must complete within 10 minutes or will fail
 //   go run withContext.go -b mybucket -k myKey -d 10m < myfile.txt
func example_s3(bucket string, key string, timeout time.Duration) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := s3.New(sess)

 	// Create a context with a timeout that will abort the upload if it takes more than the passed in timeout.
 	ctx := context.Background()
 	var cancelFn func()
 	if timeout > 0 {
 		ctx, cancelFn = context.WithTimeout(ctx, timeout)
 	}
 	
 	defer cancelFn() // Ensure the context is canceled to prevent leaking.

 	// Uploads the object to S3. The Context will interrupt the request if the timeout expires.
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

func cloudwatchMetrics(metricName string, metricId string, stat string, startTime Time, endTime Time, period int64) {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)

 	//endTime := time.Now()
	//duration, _ := time.ParseDuration("-5m")
	//startTime := endTime.Add(duration)
	//metricname := "ClusterUsedSpace"
	//metricid := "m1"
	//period := int64(60)
	//stat := "Average"
	namespace := "AWS/EC2"
	metricDimName := "InstanceId"
	metricDimValue := "i-0706dcb2c513b981c"

	query := &cloudwatch.MetricDataQuery{
		Id: &metricId,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &namespace,
				MetricName: &metricName,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  &metricDimName,
						Value: &metricDimValue,
					},
				},
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

	for _, metricdata := range resp.MetricDataResults {
		fmt.Println(*metricdata.Id)
		for index, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (*metricdata.Timestamps[index]).String(), *metricdata.Values[index])
		}
	}
}

 func main() {
 	var bucket, key string
 	var timeout time.Duration

 	flag.StringVar(&bucket, "b", "", "Bucket name.")
 	flag.StringVar(&key, "k", "", "Object key name.")
 	flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
 	flag.Parse()

 	//example_s3(bucket, key, timeout)
 	//fmt.Printf("successfully uploaded file to %s/%s\n", bucket, key)
 	startTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:30:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-10-17T12:40:00+02:00")
 	cloudwatchMetrics("CPUUtilization", "cpu1", "Average", startTime, endTime, 300)
 }
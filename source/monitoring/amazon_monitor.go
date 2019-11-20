 package main

import (
 	"fmt"
 	"log"
 	"time"
 	"strings"
 	"github.com/aws/aws-sdk-go/aws"
 	"github.com/aws/aws-sdk-go/aws/session"
 	"github.com/aws/aws-sdk-go/service/cloudwatch"
 	"progettoSDCC/source/utility"
 )

const (
 	awsRegion = "us-east-1"
 	EC2MetricJsonPath = "../../configuration/metrics_ec2.json"
 	EC2InstPath = "../../configuration/ec2_inst.json"
 	S3MetricPath = "../../configuration/metrics_s3.json"
 	S3BucketPath = "../../configuration/s3_buckets.json"
 )

type AWSMetric struct {
	Namespace string
 	Name string
 	Dimensions []*cloudwatch.Dimension
 }

 type AWSMetricJson struct{
 	AWSMetric AWSMetric
 	Monitoring bool
 	DimensionNames []string
 	DimensionValues []string
 }

type EC2Inst []string

func cloudwatchClient(awsRegion string) *cloudwatch.CloudWatch{
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)
 	return svc
}

func importMetrics(metricsJ []AWSMetricJson) []AWSMetric {
	res := make([]AWSMetric, 0)
	for _, metricJ := range metricsJ {
		if metricJ.Monitoring == true {
			dimensions := make([]*cloudwatch.Dimension, len(metricJ.DimensionValues))
			for j,_ := range metricJ.DimensionValues {
				dimensions[j] = &cloudwatch.Dimension{
					Name : &metricJ.DimensionNames[j],
					Value : &metricJ.DimensionValues[j],
				}
			}
			metricJ.AWSMetric.Dimensions = dimensions

			res = append(res, metricJ.AWSMetric)
		}
	}
	return res
}

func loadDimensionsSameName(metrics []AWSMetric, dimensionName string, dimensionValues []string) []AWSMetric{
	for i, _ := range metrics {
		for j,_ := range dimensionValues {
			dimension := &cloudwatch.Dimension{
				Name : &dimensionName,
				Value : &dimensionValues[j],
			}
			metrics[i].Dimensions = append(metrics[i].Dimensions, dimension)
		}
	}
	return metrics
}

func loadDimensions(metrics []AWSMetric, dimensionNames []string, dimensionValues []string) []AWSMetric{
	for i, _ := range metrics {
		for j,_ := range dimensionNames {
			dimension := &cloudwatch.Dimension{
				Name : &dimensionNames[j],
				Value : &dimensionValues[j],
			}
			metrics[i].Dimensions = append(metrics[i].Dimensions, dimension)
		}
	}
	return metrics
}

func cloudwatchPrintMetrics(results []*cloudwatch.MetricDataResult) {
	for _, metricdata := range results {
	fmt.Println(*metricdata.Label)
	for j, _ := range metricdata.Timestamps {
		fmt.Printf("%v %v\n", (*metricdata.Timestamps[j]).String(), *metricdata.Values[j])
		}
	} 
}

func cloudwatchGetMetrics(svc *cloudwatch.CloudWatch, metrics []AWSMetric, 
							stat string, startTime time.Time, endTime time.Time, 
							period int64) []*cloudwatch.MetricDataResult {
	query := make([]*cloudwatch.MetricDataQuery, len(metrics))
	for i := 0; i < len(metrics); i++ {
		query[i] = &cloudwatch.MetricDataQuery{
			Id: aws.String(strings.ToLower(metrics[i].Name)),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  &metrics[i].Namespace,
					MetricName: &metrics[i].Name,
					Dimensions: metrics[i].Dimensions,
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
		log.Fatal("Error in GetMetricData", err)
	}
	return resp.MetricDataResults;
}

 func main() {
 	startTime, _ := time.Parse(time.RFC3339, "2019-11-09T15:35:00+02:00")
 	endTime, _ := time.Parse(time.RFC3339, "2019-11-09T16:00:00+02:00")
 	var instanceIds, bucketNames []string
 	var ec2MetricsJ, s3MetricsJ []AWSMetricJson

	utility.ImportJson(EC2MetricJsonPath, &ec2MetricsJ)
 	utility.ImportJson(EC2InstPath, &instanceIds)
 	utility.ImportJson(S3MetricPath, &s3MetricsJ)
 	utility.ImportJson(S3BucketPath, &bucketNames)

 	ec2Metrics := importMetrics(ec2MetricsJ)
 	s3Metrics := importMetrics(s3MetricsJ)

 	ec2Metrics = loadDimensionsSameName(ec2Metrics, "InstanceId", instanceIds)
 	s3Metrics = loadDimensionsSameName(s3Metrics, "BucketName", bucketNames)

 	//fmt.Println(s3Metrics)

 	svc := cloudwatchClient(awsRegion)
 	
	ec2Data := cloudwatchGetMetrics(svc, ec2Metrics, "Average", startTime, endTime, 300)
	cloudwatchPrintMetrics(ec2Data)
	s3Data := cloudwatchGetMetrics(svc, s3Metrics, "Average", startTime, endTime, 300)
	cloudwatchPrintMetrics(s3Data)	
 }
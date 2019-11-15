package aws_monitor

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
 	StatPath = "../../configuration/monitoring_stat.json"
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

 type AWSStat struct{
 	Stat string
 	Period int64
 }

 type AWSMonitor struct {
 	client *cloudwatch.CloudWatch
 	stat AWSStat
 	metrics []AWSMetric
 }


func cloudwatchClient(awsRegion string) *cloudwatch.CloudWatch{
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)
 	return svc
}

func importMetrics(metricsJ []AWSMetricJson) ([]AWSMetric, [][] *cloudwatch.Dimension) {
	metrics := make([]AWSMetric, 0)
	dimensions := make([][]*cloudwatch.Dimension, len(metricsJ))
	for i, metricJ := range metricsJ {
		if metricJ.Monitoring == true {
			for j,_ := range metricJ.DimensionValues {
				dimension := &cloudwatch.Dimension{
					Name : &metricJ.DimensionNames[j],
					Value : &metricJ.DimensionValues[j],
				}
				dimensions[i] = append(dimensions[i], dimension)
			}
			metrics = append(metrics, metricJ.AWSMetric)
		}
	}
	return metrics, dimensions
}

func appendDimensions(metrics []AWSMetric, dimensions [][] *cloudwatch.Dimension) []AWSMetric {
	for i,_ := range dimensions {
		metrics[i].Dimensions = append(metrics[i].Dimensions, dimensions[i]...)
	}
	return metrics
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

func PrintMetrics(results []*cloudwatch.MetricDataResult) {
	for _, metricdata := range results {
	fmt.Println(*metricdata.Label)
	for j, _ := range metricdata.Timestamps {
		fmt.Printf("%v %v\n", (*metricdata.Timestamps[j]).String(), *metricdata.Values[j])
		}
	} 
}

func (monitor *AWSMonitor) GetMetrics(startTime time.Time, endTime time.Time) []*cloudwatch.MetricDataResult {
	query := make([]*cloudwatch.MetricDataQuery, len(monitor.metrics))
	for i,_ := range monitor.metrics {
		query[i] = &cloudwatch.MetricDataQuery{
			Id: aws.String(strings.ToLower(monitor.metrics[i].Name)),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Namespace:  &monitor.metrics[i].Namespace,
					MetricName: &monitor.metrics[i].Name,
					Dimensions: monitor.metrics[i].Dimensions,
				},
				Period: &monitor.stat.Period,
				Stat:   &monitor.stat.Stat,
			},
		}
	}

	resp, err := monitor.client.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: query,
	})

	if err != nil {
		log.Fatal("Error in GetMetricData", err)
	}
	return resp.MetricDataResults;
}

func New() AWSMonitor{
	var ec2MetricsJ, s3MetricsJ []AWSMetricJson
	var instanceIds []string
	var stat AWSStat

	utility.ImportJson(EC2MetricJsonPath, &ec2MetricsJ)
 	utility.ImportJson(EC2InstPath, &instanceIds)
 	utility.ImportJson(S3MetricPath, &s3MetricsJ)
 	utility.ImportJson(StatPath, &stat)

 	ec2Metrics, _ := importMetrics(ec2MetricsJ)
 	s3Metrics, s3Dim := importMetrics(s3MetricsJ)
 	ec2Metrics = loadDimensionsSameName(ec2Metrics, "InstanceId", instanceIds)
 	s3Metrics = appendDimensions(s3Metrics, s3Dim)
 	metrics := append(ec2Metrics, s3Metrics...)

	svc := cloudwatchClient(awsRegion)

	return AWSMonitor{svc, stat, metrics}
}
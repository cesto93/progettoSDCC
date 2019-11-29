package monitor

import (
 	"fmt"
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
 	EC2InstPath = "../../configuration/generated/ec2_inst.json"
 	S3MetricPath = "../../configuration/metrics_s3.json"
 	StatPath = "../../configuration/monitoring_stat.json"
 )

type AwsMonitor struct {
 	client *cloudwatch.CloudWatch
 	stat AwsStat
 	metrics []AwsMetric
}

//Used by json
type AwsMetric struct {
	Namespace string
 	Name string
 	Dimensions []*cloudwatch.Dimension
}

//Used by json
type AwsMetricJson struct{
 	AWSMetric AwsMetric
 	Monitoring bool
 	DimensionNames []string
 	DimensionValues []string
}

//Used by json
type AwsStat struct{
 	Stat string
 	Period int64
}

func cloudwatchClient(awsRegion string) *cloudwatch.CloudWatch{
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(awsRegion), }))
 	svc := cloudwatch.New(sess)
 	return svc
}

func importMetrics(metricsJ []AwsMetricJson) ([]AwsMetric, [][] *cloudwatch.Dimension) {
	metrics := make([]AwsMetric, 0)
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

func appendDimensions(metrics []AwsMetric, dimensions [][] *cloudwatch.Dimension) []AwsMetric {
	for i,_ := range dimensions {
		metrics[i].Dimensions = append(metrics[i].Dimensions, dimensions[i]...)
	}
	return metrics
}

func loadDimensionsSameName(metrics []AwsMetric, dimensionName string, dimensionValues []string) []AwsMetric{
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

//UNUSED
func loadDimensions(metrics []AwsMetric, dimensionNames []string, dimensionValues []string) []AwsMetric{
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

//DEBUG
func printAwsMetrics(results []*cloudwatch.MetricDataResult) {
	for _, metricdata := range results {
		fmt.Println(*metricdata.Label)
		for j, _ := range metricdata.Timestamps {
			fmt.Printf("%v %v\n", (*metricdata.Timestamps[j]).String(), *metricdata.Values[j])
		}
	} 
}

func (monitor *AwsMonitor) getMetrics(startTime time.Time, endTime time.Time) ([]*cloudwatch.MetricDataResult, error) {
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
		return nil, fmt.Errorf("Error in GetMetricData, %v", err)
	}
	return resp.MetricDataResults, nil;
}

func (monitor *AwsMonitor) GetMetrics(startTime time.Time, endTime time.Time) ([]MetricData, error) {
	awsRes, err := monitor.getMetrics(startTime, endTime)
	res := make([]MetricData, len(awsRes))

	if err != nil {
		return nil, err
	}
	
	for i, metricdata := range awsRes {
		res[i].Label = *metricdata.Label
		res[i].Values = make([]float64, len(metricdata.Timestamps))
		res[i].Timestamps = make([]time.Time, len(metricdata.Timestamps))
		for j, _ := range metricdata.Timestamps {
			res[i].Values[j] = *metricdata.Values[j]
			res[i].Timestamps[j] = *metricdata.Timestamps[j]
		}
	}
	return res, nil 
}

func NewAws() *AwsMonitor {
	var ec2MetricsJ, s3MetricsJ []AwsMetricJson
	var instanceIds []string
	var stat AwsStat

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

	return &AwsMonitor{svc, stat, metrics}
}
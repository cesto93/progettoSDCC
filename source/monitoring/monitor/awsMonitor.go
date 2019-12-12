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
 )

type AwsMonitor struct {
 	client *cloudwatch.CloudWatch
 	stat string
 	period int64
 	metrics []AwsMetric
}

//Used by json
type AwsMetric struct {
	Namespace string
 	Name string
 	Dimensions []*cloudwatch.Dimension
 	Unit string
}

//Used by json
type AwsMetricJson struct{
 	AWSMetric AwsMetric
 	Monitoring bool
 	DimensionNames []string
 	DimensionValues []string
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

/*//UNUSED
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
}*/

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
				Period: &monitor.period,
				Stat:   &monitor.stat,
				Unit: &monitor.metrics[i].Unit,
			},
		}
	}

	fmt.Printf("query: %v\n", query)

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
	res := make([]MetricData, 0)

	if err != nil {
		return nil, err
	}
	
	for _, metricdata := range awsRes {
		if len(metricdata.Values) != 0 { 
			var data MetricData 
			s := strings.Split(*metricdata.Label, " ")
			data.TagName = s[0]
			data.TagValue = s[1]
			data.Label = s[2]
			data.Values = make([]interface{}, len(metricdata.Timestamps))
			data.Timestamps = make([]time.Time, len(metricdata.Timestamps))
			for j, _ := range metricdata.Values {
				data.Values[j] = *metricdata.Values[j]
				data.Timestamps[j] = *metricdata.Timestamps[j]
			}
			res = append(res, data)
		}
	}
	return res, nil 
}

func NewAws(ec2MetricJsonPath string, ec2InstPath string, s3MetricPath string,  stat string, period int64) *AwsMonitor {
	var ec2MetricsJ, s3MetricsJ []AwsMetricJson
	var instanceIds []string
	//var stat AwsStat

	utility.ImportJson(ec2MetricJsonPath, &ec2MetricsJ)
 	utility.ImportJson(ec2InstPath, &instanceIds)
 	utility.ImportJson(s3MetricPath, &s3MetricsJ)
 	//utility.ImportJson(statPath, &stat)

 	ec2Metrics, _ := importMetrics(ec2MetricsJ)
 	s3Metrics, s3Dim := importMetrics(s3MetricsJ)
 	ec2Metrics = loadDimensionsSameName(ec2Metrics, "InstanceId", instanceIds)
 	s3Metrics = appendDimensions(s3Metrics, s3Dim)
 	metrics := append(ec2Metrics, s3Metrics...)

	svc := cloudwatchClient(awsRegion)

	return &AwsMonitor{svc, stat, period, metrics}
}
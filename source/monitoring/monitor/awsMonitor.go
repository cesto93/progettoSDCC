package monitor

import (
 	"fmt"
 	"time"
 	"strconv"
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

func loadDimensionSameName(metrics []AwsMetric, dimensionName string, dimensionValues string) []AwsMetric{
	for i, _ := range metrics {
		dimension := &cloudwatch.Dimension{
			Name : &dimensionName,
			Value : &dimensionValues,
		}
		metrics[i].Dimensions = []*cloudwatch.Dimension{dimension}
	}
	return metrics
}

func metricsSetInstance(metrics []AwsMetric, dimensionName string, instanceIds []string) []AwsMetric{
	var allmetrics []AwsMetric
	for i := range instanceIds {
		allmetrics = append(allmetrics, loadDimensionSameName(metrics, dimensionName, instanceIds[i])...)
	}
	return allmetrics
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
		//id := *monitor.metrics[i].Dimensions[0].Value
		query[i] = &cloudwatch.MetricDataQuery{
			Id: aws.String("id_" + strconv.Itoa(i)),
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
			id,_ := strconv.Atoi((*metricdata.Id)[3:])
			label := strings.Split(*metricdata.Label, " ")
			data.TagName = "AWS/EC2"
			data.TagValue = *monitor.metrics[id].Dimensions[0].Value
			data.Label = label[len(label)-1]
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

func NewAws(ec2MetricJsonPath string, ec2InstPath string,  stat string, period int64) *AwsMonitor {
	var ec2MetricsJ []AwsMetricJson
	var instanceIds []string

	utility.CheckError(utility.ImportJson(ec2MetricJsonPath, &ec2MetricsJ))
 	utility.CheckError(utility.ImportJson(ec2InstPath, &instanceIds))

 	ec2Metrics, _ := importMetrics(ec2MetricsJ)
 	ec2Metrics = metricsSetInstance(ec2Metrics, "InstanceId", instanceIds)
	svc := cloudwatchClient(awsRegion)

	return &AwsMonitor{svc, stat, period, ec2Metrics}
}

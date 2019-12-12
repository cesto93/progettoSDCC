package main

import (
 	"time"
 	"fmt"
 	"flag"
 	"progettoSDCC/source/monitoring/monitor"
 	"progettoSDCC/source/monitoring/zookeeper"
 	"progettoSDCC/source/monitoring/restarter"
 	"progettoSDCC/source/monitoring/db"
 	"progettoSDCC/source/utility"
 	"progettoSDCC/source/appMetrics"
 )

 const (
 	sessionTimeout = 10
 	monitorIntervalSeconds = 300
 	restartIntervalSecond = 5

 	GCEprojectIDPath = "../configuration/generated/gce_project_id.json"

 	dbAddrPath = "../configuration/generated/db_addr.json"
 	dbName = "mydb"

 	zkServersIpPath = "../configuration/generated/zk_servers_addrs.json"
 	zkAgentPath = "../configuration/generated/zk_agent.json"
 	idMonitorPath = "../configuration/generated/id_monitor.json"
 	aliveNodePath = "/alive"

 	EC2MetricJsonPath = "../configuration/metrics_ec2.json"
 	EC2InstPath = "../configuration/generated/ec2_inst.json"
 	S3MetricPath = "../configuration/metrics_s3.json"
 	StatPath = "../configuration/monitoring_stat.json"

 	GcloudMetricsJsonPath = "../configuration/metrics_gce.json"
    InstancesJsonPath = "../configuration/generated/instances_ids.json"

    PrometheusMetricsJsonPath = "../configuration/metrics_prometheus.json"

    AppmetricsJsonPath = "../log/app_metrics.json"
 )

 func saveMetrics(monitorBridge monitor.MonitorBridge, dbBridge *db.DbBridge, start time.Time, end time.Time) {
 	data, err := monitorBridge.GetMetrics(start, end)
 	utility.CheckError(err)
 	printMetrics(data, start, end)
 	err = dbBridge.SaveMetrics(data)
 	for err != nil {
 		err = dbBridge.SaveMetrics(data)
 	}
 }

 func saveAppMetrics(path string, dbBridge *db.DbBridge) {
 	var metrics []monitor.MetricData
 	var wordElaboratedApplication, elaborationTime, throughPutApplication, workers []interface{}
	var timestamp []time.Time

 	raws, err := appMetrics.ReadApplicationMetrics(path)
 	utility.CheckError(err)
 	for _, raw := range raws {
 		wordElaboratedApplication = append(wordElaboratedApplication, raw.Metrics.WordElaboratedApplication)
 		elaborationTime = append(elaborationTime, raw.Metrics.ElaborationTime)
 		throughPutApplication = append(throughPutApplication, raw.Metrics.ThroughPutApplication)
 		workers = append(workers, raw.Metrics.Workers)
 		timestamp = append(timestamp, raw.Timestamp)
 	}

 	metrics = append(metrics, 
 	monitor.MetricData  {
		Label : "WordElaboratedApplication",
		TagName : "App",
		TagValue : "Word_Count",
		Timestamps: timestamp,
		Values : wordElaboratedApplication,
	} )

	metrics = append(metrics, 
 	monitor.MetricData  {
		Label : "ElaborationTime",
		TagName : "App",
		TagValue : "Word_Count",
		Timestamps: timestamp,
		Values : elaborationTime,
	} )

	metrics = append(metrics, 
 	monitor.MetricData  {
		Label : "ThroughPutApplication",
		TagName : "App",
		TagValue : "Word_Count",
		Timestamps: timestamp,
		Values : throughPutApplication,
	} )

	metrics = append(metrics, 
 	monitor.MetricData  {
		Label : "Workers",
		TagName : "App",
		TagValue : "Word_Count",
		Timestamps: timestamp,
		Values : workers,
	} )

 	printMetrics(metrics, timestamp[0], timestamp[len(timestamp)-1])
 	err = dbBridge.SaveMetrics(metrics)
 	for err != nil {
 		err = dbBridge.SaveMetrics(metrics)
 	}
 }

 func checkMembersDead(zkBridge *zookeeper.ZookeeperBridge, id string) {
 	for {
 		zkBridge.CheckMemberDead(id)
 	}
 }

 func printMetrics(results []monitor.MetricData, start time.Time, end time.Time) {
 	fmt.Printf("Request from %v to %v: \n", start, end)
	for _, metricdata := range results {
		fmt.Printf("%v %v : %v\n", metricdata.Label, metricdata.TagName, metricdata.TagValue)
		for j, _ := range metricdata.Timestamps {
			fmt.Printf("%v value: %v\n", (metricdata.Timestamps[j]).String(), metricdata.Values[j])
		}
	} 
}

func main() {
	var zkServerAddresses, members []string
	var dbAddr string
	var monitorBridge, monitorPrometheus monitor.MonitorBridge
	var myRestarter restarter.Restarter
	var aws,tryed bool
	var index, next int

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")
	flag.Parse()

	// db conf
	err := utility.ImportJson(dbAddrPath, &dbAddr)
	utility.CheckError(err)
	dbBridge:= db.NewDb(dbAddr, dbName)

	// zk conf
 	err = utility.ImportJson(zkAgentPath, &members)
 	utility.CheckError(err)
 	err = utility.ImportJson(zkServersIpPath, &zkServerAddresses)
 	utility.CheckError(err)
 	err = utility.ImportJson(idMonitorPath, &index)
 	utility.CheckError(err)
 	next = (index + 1) % len(members) //this is the id of agent to restart if crash
 	fmt.Println(zkServerAddresses) //less than 3 servers dosen't make zookeeper fault tolerant

 	zkBridge, err := zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath)
 	for err != nil {
 		fmt.Println(err)
 		time.Sleep(time.Second * restartIntervalSecond)
 		zkBridge, err = zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath)
 	}
 	
 	err = zkBridge.RegisterMember(members[index], "info")
 	for err != nil {
 		fmt.Println(err)
 		time.Sleep(time.Second * restartIntervalSecond)
 		err = zkBridge.RegisterMember(members[index], "info")
 	}
 	go checkMembersDead(zkBridge, members[next])

 	if (aws) {
 		monitorBridge = monitor.NewAws(EC2MetricJsonPath, EC2InstPath, S3MetricPath, "Average", monitorIntervalSeconds)
 		myRestarter = restarter.NewAws()
 	} else {
 		var GCEprojectID string
 		err = utility.ImportJson(GCEprojectIDPath, &GCEprojectID)
 		utility.CheckError(err)
 		monitorBridge = monitor.NewGce(GCEprojectID, GcloudMetricsJsonPath, InstancesJsonPath)
 		myRestarter = restarter.NewGce()
 	}
 	monitorPrometheus = monitor.NewPrometheus(PrometheusMetricsJsonPath)

	monitorInterval := monitorIntervalSeconds * time.Second
	now := time.Now()
	lastMeasure := now.Add(-monitorInterval)
	nextMeasure := now

	for {
		time.Sleep(time.Second * restartIntervalSecond)
		now = time.Now()
		if zkBridge.IsDead != false {
			fmt.Println("This is is dead: " + members[next])
			tryed, err = myRestarter.Restart(members[next])
			utility.CheckErrorNonFatal(err)

			if tryed {
				fmt.Println("Tried to restart")
			}
		}
		if (now.After(nextMeasure)) {
			saveMetrics(monitorBridge, dbBridge, lastMeasure.UTC(), nextMeasure.UTC())
			saveMetrics(monitorPrometheus, dbBridge, lastMeasure.UTC(), nextMeasure.UTC())
			saveAppMetrics(AppmetricsJsonPath, dbBridge)
			lastMeasure = lastMeasure.Add(monitorInterval)
			nextMeasure = nextMeasure.Add(monitorInterval)
		}
	}
 }
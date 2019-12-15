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

 	dbAddrPath = "../configuration/generated/db_addr.json"
 	dbName = "mydb"

 	zkServersIpPath = "../configuration/generated/zk_servers_addrs.json"
 	zkAgentPath = "../configuration/generated/zk_agent.json"
 	idMonitorPath = "../configuration/generated/id_monitor.json"
 	aliveNodePath = "/alive"

 	EC2MetricJsonPath = "../configuration/metrics_ec2.json"
 	EC2InstPath = "../configuration/generated/ec2_inst.json"
 	StatPath = "../configuration/monitoring_stat.json"

 	GCEprojectIDPath = "../configuration/generated/gce_project_id.json"
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
 	metrics, err := appMetrics.ReadApplicationMetrics(path)
 	utility.CheckError(err)
 	if (metrics != nil) {
 		printMetrics(metrics, metrics[0].Timestamps[0], metrics[0].Timestamps[0])
 		err = dbBridge.SaveMetrics(metrics)
 		for err != nil {
 			err = dbBridge.SaveMetrics(metrics)
 		}
 	}
 }

 func checkMembersDead(zkBridge *zookeeper.ZookeeperBridge, id string) {
 	for {
 		err := zkBridge.CheckMemberDead(id)
 		utility.CheckError(err)
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

func recoverState(dbBridge *db.DbBridge, monitorBridge monitor.MonitorBridge, monitorInterval time.Duration) {
	start, err := dbBridge.GetLastTimestamp("Up")
	i := 0
	if err != nil && i < 5 {
		start, err = dbBridge.GetLastTimestamp("Up")
	}
	utility.CheckErrorNonFatal(err)
	now := time.Now()
	end := start.Add(monitorInterval)
	for ; start.After(now); start.Add(monitorInterval) {
		saveMetrics(monitorBridge, dbBridge, *start, end)
		end.Add(monitorIntervalSeconds)
	}
}

func main() {
	var zkServerAddresses, members []string
	var dbAddr string
	var monitorBridge, monitorPrometheus monitor.MonitorBridge
	var myRestarter restarter.Restarter
	var aws bool
	var index, next int
	var start, end time.Time

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")
	flag.Parse()
	
	//wait interval
	monitorInterval := monitorIntervalSeconds * time.Second
	restartInterval := time.Second * restartIntervalSecond

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
 		time.Sleep(restartInterval)
 		zkBridge, err = zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath)
 	}
 	
 	err = zkBridge.RegisterMember(members[index], "info")
 	for err != nil {
 		fmt.Println(err)
 		time.Sleep(time.Second * restartIntervalSecond)
 		err = zkBridge.RegisterMember(members[index], "info")
 	}
 	go checkMembersDead(zkBridge, members[next])

 	now := time.Now().Truncate(monitorInterval)
	lastMeasure := now.Add(-monitorInterval)
	nextMeasure := now

 	if (aws) {
 		monitorBridge = monitor.NewAws(EC2MetricJsonPath, EC2InstPath, "Average", monitorIntervalSeconds)
 		myRestarter = restarter.NewAws()
 		awsDelay := 5 * time.Minute
 		start = lastMeasure.Add(-awsDelay)
 		end = nextMeasure.Add(-awsDelay)
 	} else {
 		var GCEprojectID string
 		err = utility.ImportJson(GCEprojectIDPath, &GCEprojectID)
 		utility.CheckError(err)
 		monitorBridge = monitor.NewGce(GCEprojectID, GcloudMetricsJsonPath, InstancesJsonPath)
 		myRestarter = restarter.NewGce(GCEprojectID)
 		start = lastMeasure
 		end = nextMeasure
 	}
 	monitorPrometheus = monitor.NewPrometheus(PrometheusMetricsJsonPath)
 	tryed := false
 	
 	fmt.Printf("Starting agent %s\n that observ %s\n", members[index], members[next])
 	recoverState(dbBridge, monitorBridge, monitorInterval)

	for {
		time.Sleep(restartInterval)
		now = time.Now()
		if zkBridge.IsDead != false {
			fmt.Println("This is is dead: " + members[next])
			tryed, err = myRestarter.Restart(members[next])
			utility.CheckErrorNonFatal(err)

			if tryed {
				fmt.Println("Tried to restart")
				time.Sleep(time.Second * restartIntervalSecond)
				zkBridge.IsDead = false
				tryed = false
			}
		}
		if (now.After(nextMeasure)) {
			saveMetrics(monitorBridge, dbBridge, start.UTC(), end.UTC())
			saveMetrics(monitorPrometheus, dbBridge, lastMeasure.UTC(), nextMeasure.UTC())
			saveAppMetrics(AppmetricsJsonPath, dbBridge)
			lastMeasure = lastMeasure.Add(monitorInterval)
			nextMeasure = nextMeasure.Add(monitorInterval)
			start = start.Add(monitorInterval)
			end = end.Add(monitorInterval)
		}
	}
 }

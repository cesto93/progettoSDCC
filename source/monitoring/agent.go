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

 	retryDB = 5
 	retryZK = 10

 	dbAddrPath = "../configuration/generated/db_addr.json"
 	dbName = "mydb"

 	zkServersIpPath = "../configuration/generated/zk_servers_addrs.json"
 	zkAgentPath = "../configuration/generated/zk_agent.json"
 	idMonitorPath = "../configuration/generated/id_monitor.json"
 	aliveNodePath = "/alive"

 	EC2MetricJsonPath = "../configuration/metrics_ec2.json"
 	EC2InstPath = "../configuration/generated/ec2_inst.json"

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
 	for i:=0; err != nil && i < retryDB; i++  {
		utility.CheckErrorNonFatal(err)
 		err = dbBridge.SaveMetrics(data)
 		time.Sleep(time.Second * 3)
 	}
 }

 func saveAppMetrics(path string, dbBridge *db.DbBridge) {
 	metrics, err := appMetrics.ReadApplicationMetrics(path)
 	utility.CheckError(err)
 	if (metrics != nil) {
 		printMetrics(metrics, metrics[0].Timestamps[0], metrics[0].Timestamps[0])
 		err = dbBridge.SaveMetrics(metrics)
 		for i:=0; err != nil && i < retryDB; i++  {
			utility.CheckErrorNonFatal(err)
 			err = dbBridge.SaveMetrics(metrics)
			time.Sleep(time.Second * 3)
 		}
 	}
 }

 func restoreMembersDead(zkBridge *zookeeper.ZookeeperBridge, id string, myRestarter restarter.Restarter, 
							restartInterval time.Duration) {
	tryed := false 
 	for {
 		err := zkBridge.CheckMemberDead(id)
 		utility.CheckError(err)
 		if zkBridge.IsDead != false {
			fmt.Println("This is is dead: " + id)
			tryed, err = myRestarter.Restart(id)
			utility.CheckErrorNonFatal(err)

			if tryed {
				fmt.Println("Tried to restart")
				time.Sleep(restartInterval)
				zkBridge.IsDead = false
				tryed = false
			}
		}
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
	for i := 0; err != nil && i < retryDB; i++ {
		utility.CheckErrorNonFatal(err)
		start, err = dbBridge.GetLastTimestamp("Up")
		time.Sleep(time.Second * 3)
	}
	if err != nil {
		fmt.Println("Cannot recover state")
	}
	end := time.Now().Truncate(monitorInterval)
	saveMetrics(monitorBridge, dbBridge, *start, end)
}

func main() {
	var zkServerAddresses, members []string
	var dbAddr string
	var monitorBridge, monitorPrometheus monitor.MonitorBridge
	var myRestarter restarter.Restarter
	var aws, disableRecover bool
	var index, next int
	var start, end time.Time

	flag.BoolVar(&aws, "aws", false, "Specify the aws monitor")
	flag.BoolVar(&aws, "disableRecover", false, "Disable the recovery state function")
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
 	for i:=0; err != nil && i < retryZK; i++  {
 		utility.CheckErrorNonFatal(err)
 		zkBridge, err = zookeeper.New(zkServerAddresses, time.Second * sessionTimeout, aliveNodePath)
 		time.Sleep(time.Second * 3)
 	}
 	
 	err = zkBridge.RegisterMember(members[index], "info")
 	for err != nil {
 		utility.CheckErrorNonFatal(err)
 		err = zkBridge.RegisterMember(members[index], "info")
 		time.Sleep(time.Second * 3)
 	}

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
 	
 	go restoreMembersDead(zkBridge, members[next], myRestarter, restartInterval)
 	monitorPrometheus = monitor.NewPrometheus(PrometheusMetricsJsonPath)
 	
 	fmt.Printf("Starting agent %s\n that observ %s\n", members[index], members[next])
 	if disableRecover {
		recoverState(dbBridge, monitorBridge, monitorInterval)
	}

	for {
		now = time.Now()
		if (now.After(nextMeasure)) {
			saveMetrics(monitorBridge, dbBridge, start.UTC(), end.UTC())
			saveMetrics(monitorPrometheus, dbBridge, lastMeasure.UTC(), nextMeasure.UTC())
			saveAppMetrics(AppmetricsJsonPath, dbBridge)
			lastMeasure = lastMeasure.Add(monitorInterval)
			nextMeasure = nextMeasure.Add(monitorInterval)
			start = start.Add(monitorInterval)
			end = end.Add(monitorInterval)
		}
		time.Sleep(restartInterval)
	}
 }
